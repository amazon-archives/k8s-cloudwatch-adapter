package provider

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	"k8s.io/metrics/pkg/apis/external_metrics"
)

func (p *cloudwatchProvider) GetExternalMetric(namespace string, metricSelector labels.Selector, info provider.ExternalMetricInfo) (*external_metrics.ExternalMetricValueList, error) {
	// Note:
	//		metric name and namespace is used to lookup for the CRD which contains configuration to
	//		call cloudwatch if not found then ignored and label selector is parsed for all the metrics
	klog.V(0).Infof("Received request for namespace: %s, metric name: %s, metric selectors: %s", namespace, info.Metric, metricSelector.String())

	_, selectable := metricSelector.Requirements()
	if !selectable {
		return nil, errors.NewBadRequest("label is set to not selectable. this should not happen")
	}

	externalRequest, found := p.metricCache.GetExternalMetric(namespace, info.Metric)
	if !found {
		return nil, errors.NewBadRequest("no metric query found")
	}

	metricValue, err := p.cwManager.QueryCloudWatch(externalRequest)
	if err != nil {
		klog.Errorf("bad request: %v", err)
		return nil, errors.NewBadRequest(err.Error())
	}

	var quantity resource.Quantity
	if len(metricValue) == 0 || len(metricValue[0].Values) == 0 {
		quantity = *resource.NewMilliQuantity(0, resource.DecimalSI)
	} else {
		quantity = *resource.NewQuantity(int64(aws.Float64Value(metricValue[0].Values[0])), resource.DecimalSI)
	}
	externalmetric := external_metrics.ExternalMetricValue{
		MetricName: info.Metric,
		Value:      quantity,
		Timestamp:  metav1.Now(),
	}

	matchingMetrics := []external_metrics.ExternalMetricValue{}
	matchingMetrics = append(matchingMetrics, externalmetric)

	return &external_metrics.ExternalMetricValueList{
		Items: matchingMetrics,
	}, nil
}

func (p *cloudwatchProvider) ListAllExternalMetrics() []provider.ExternalMetricInfo {
	p.valuesLock.RLock()
	defer p.valuesLock.RUnlock()

	// not implemented yet
	externalMetricsInfo := []provider.ExternalMetricInfo{}
	for _, name := range p.metricCache.ListMetricNames() {
		// only process if name is non-empty
		if name != "" {
			info := provider.ExternalMetricInfo{
				Metric: name,
			}
			externalMetricsInfo = append(externalMetricsInfo, info)
		}
	}
	return externalMetricsInfo
}
