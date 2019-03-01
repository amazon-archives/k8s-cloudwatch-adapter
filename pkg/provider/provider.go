package provider

import (
	"sync"
	"time"

	"github.com/golang/glog"

	apierr "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/metrics/pkg/apis/custom_metrics"

	"github.com/chankh/k8s-cloudwatch-adapter/pkg/aws"
	"github.com/chankh/k8s-cloudwatch-adapter/pkg/config"
	"github.com/chankh/k8s-cloudwatch-adapter/pkg/metriccache"
	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider/helpers"
)

var nsGroupResource = schema.GroupResource{Resource: "namespaces"}

// CustomMetricResource wraps provider.CustomMetricInfo in a struct which stores the Name and Namespace of the resource
// So that we can accurately store and retrieve the metric as if this were an actual metrics server.
type CustomMetricResource struct {
	provider.CustomMetricInfo
	types.NamespacedName
}

// cloudwatchProvider is a implementation of provider.MetricsProvider for CloudWatch
type cloudwatchProvider struct {
	client   dynamic.Interface
	mapper   apimeta.RESTMapper
	cwClient aws.Client

	valuesLock  sync.RWMutex
	values      map[CustomMetricResource]resource.Quantity
	series      []config.MetricSeriesConfig
	metricCache *metriccache.MetricCache

	// info maps metric info to information about the corresponding series
	info map[provider.CustomMetricInfo]string

	// metrics is the list of all known metrics
	metrics []provider.CustomMetricInfo
}

// NewFakeProvider returns an instance of testingProvider, along with its restful.WebService that opens endpoints to post new fake metrics
func NewCloudWatchProvider(client dynamic.Interface, mapper apimeta.RESTMapper, cwClient aws.Client, series []config.MetricSeriesConfig, metricCache *metriccache.MetricCache) provider.MetricsProvider {
	provider := &cloudwatchProvider{
		client:      client,
		mapper:      mapper,
		cwClient:    cwClient,
		values:      make(map[CustomMetricResource]resource.Quantity),
		series:      series,
		metricCache: metricCache,
	}

	provider.updateMetrics()
	return provider
}

func (p *cloudwatchProvider) queryFor(metricName string) []config.MetricDataQuery {
	for _, s := range p.series {
		if s.Name == metricName {
			return s.Queries
		}
	}

	return nil
}

func (p *cloudwatchProvider) query(res CustomMetricResource) (resource.Quantity, bool) {
	metricName := res.CustomMetricInfo.Metric
	q := p.queryFor(metricName)
	if q == nil {
		return resource.Quantity{}, false
	}
	results, err := p.cwClient.Query(q)
	if err != nil {
		glog.Errorf("cannot get metric from CloudWatch, %v", err)
		return resource.Quantity{}, false

	}
	if len(results) == 0 || len(results[0].Values) == 0 {
		return *resource.NewMilliQuantity(0, resource.DecimalSI), true
	}
	return *resource.NewMilliQuantity(int64(results[0].Values[0]*1000), resource.DecimalSI), true
}

// valueFor is a helper function to get just the value of a specific metric
func (p *cloudwatchProvider) valueFor(info provider.CustomMetricInfo, name types.NamespacedName) (resource.Quantity, error) {
	info, _, err := info.Normalized(p.mapper)
	if err != nil {
		glog.Errorf("err: %v", err)
		return resource.Quantity{}, err
	}
	metricInfo := CustomMetricResource{
		CustomMetricInfo: info,
		NamespacedName:   name,
	}

	glog.Infof("getting value for %v", metricInfo)
	value, found := p.query(metricInfo)
	if !found {
		return resource.Quantity{}, provider.NewMetricNotFoundForError(info.GroupResource, info.Metric, name.Name)
	}

	return value, nil
}

// metricFor is a helper function which formats a value, metric, and object info into a MetricValue which can be returned by the metrics API
func (p *cloudwatchProvider) metricFor(value resource.Quantity, name types.NamespacedName, selector labels.Selector, info provider.CustomMetricInfo) (*custom_metrics.MetricValue, error) {
	objRef, err := helpers.ReferenceFor(p.mapper, name, info)
	if err != nil {
		glog.Errorf("err: %v", err)
		return nil, err
	}

	metric := &custom_metrics.MetricValue{
		DescribedObject: objRef,
		MetricName:      info.Metric,
		Timestamp:       metav1.Time{time.Now()},
		Value:           value,
	}

	return metric, nil
}

// metricsFor is a wrapper used by GetMetricBySelector to format several metrics which match a resource selector
func (p *cloudwatchProvider) metricsFor(namespace string, selector labels.Selector, info provider.CustomMetricInfo) (*custom_metrics.MetricValueList, error) {
	names, err := helpers.ListObjectNames(p.mapper, p.client, namespace, selector, info)
	if err != nil {
		glog.Errorf("err: %v", err)
		return nil, err
	}

	res := make([]custom_metrics.MetricValue, 0, len(names))
	for _, name := range names {
		namespacedName := types.NamespacedName{Name: name, Namespace: namespace}
		value, err := p.valueFor(info, namespacedName)
		if err != nil {
			glog.Errorf("err: %v", err)
			if apierr.IsNotFound(err) {
				continue
			}
			return nil, err
		}

		metric, err := p.metricFor(value, namespacedName, selector, info)
		if err != nil {
			glog.Errorf("err: %v", err)
			return nil, err
		}
		res = append(res, *metric)
	}

	return &custom_metrics.MetricValueList{
		Items: res,
	}, nil
}

func (p *cloudwatchProvider) GetMetricByName(name types.NamespacedName, info provider.CustomMetricInfo) (*custom_metrics.MetricValue, error) {
	p.valuesLock.RLock()
	defer p.valuesLock.RUnlock()

	glog.Infof("GetMetricByName - name: %v, info: %v", name, info)
	value, err := p.valueFor(info, name)
	if err != nil {
		glog.Errorf("err: %v", err)
		return nil, err
	}
	return p.metricFor(value, name, labels.Everything(), info)
}

func (p *cloudwatchProvider) GetMetricBySelector(namespace string, selector labels.Selector, info provider.CustomMetricInfo) (*custom_metrics.MetricValueList, error) {
	p.valuesLock.RLock()
	defer p.valuesLock.RUnlock()

	glog.Infof("GetMetricBySelector - namespace: %v, selector: %v, info: %v", namespace, selector, info)
	return p.metricsFor(namespace, selector, info)
}

func (p *cloudwatchProvider) ListAllMetrics() []provider.CustomMetricInfo {
	p.valuesLock.RLock()
	defer p.valuesLock.RUnlock()

	return p.metrics
}

func (p *cloudwatchProvider) updateMetrics() {
	newInfo := make(map[provider.CustomMetricInfo]string)
	for _, s := range p.series {
		name := s.Name
		resource := s.Resource
		groupRes := schema.GroupResource{Group: resource.Group, Resource: resource.Resource}
		info, _, err := provider.CustomMetricInfo{
			GroupResource: groupRes,
			Namespaced:    true,
			Metric:        name,
		}.Normalized(p.mapper)

		if err != nil {
			// this is likely to show up for a lot of labels, so make it a verbose info log
			glog.V(9).Infof("unable to normalize group-resource %s from series %q, skipping: %v", groupRes.String(), name, err)
			continue
		}

		// namespace metrics aren't counted as namespaced
		if groupRes == nsGroupResource {
			info.Namespaced = false
		}

		newInfo[info] = name
	}

	newMetrics := make([]provider.CustomMetricInfo, 0, len(newInfo))
	for info := range newInfo {
		newMetrics = append(newMetrics, info)
	}

	p.valuesLock.RLock()
	defer p.valuesLock.RUnlock()

	p.info = newInfo
	p.metrics = newMetrics
}
