package controller

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	api "github.com/chankh/k8s-cloudwatch-adapter/pkg/apis/metrics/v1alpha1"
	"github.com/chankh/k8s-cloudwatch-adapter/pkg/metriccache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/chankh/k8s-cloudwatch-adapter/pkg/client/clientset/versioned/fake"
	informers "github.com/chankh/k8s-cloudwatch-adapter/pkg/client/informers/externalversions"
)

func getExternalKey(externalMetric *api.ExternalMetric) namespacedQueueItem {
	return namespacedQueueItem{
		namespaceKey: fmt.Sprintf("%s/%s", externalMetric.Namespace, externalMetric.Name),
		kind:         externalMetric.TypeMeta.Kind,
	}
}

func TestExternalMetricValueIsStored(t *testing.T) {
	var storeObjects []runtime.Object
	var externalMetricsListerCache []*api.ExternalMetric

	externalMetric := newFullExternalMetric("test")
	storeObjects = append(storeObjects, externalMetric)
	externalMetricsListerCache = append(externalMetricsListerCache, externalMetric)

	handler, metriccache := newHandler(storeObjects, externalMetricsListerCache)

	queueItem := getExternalKey(externalMetric)
	err := handler.Process(queueItem)

	if err != nil {
		t.Errorf("error after processing = %v, want %v", err, nil)
	}

	metricRequest, exists := metriccache.GetCloudWatchRequest(externalMetric.Namespace, externalMetric.Name)

	if exists == false {
		t.Errorf("exist = %v, want %v", exists, true)
	}

	validateExternalMetricResult(metricRequest, externalMetric, t)
}

func TestShouldBeAbleToStoreCustomAndExternalWithSameNameAndNamespace(t *testing.T) {
	var storeObjects []runtime.Object
	var externalMetricsListerCache []*api.ExternalMetric

	externalMetric := newFullExternalMetric("test")
	storeObjects = append(storeObjects, externalMetric)
	externalMetricsListerCache = append(externalMetricsListerCache, externalMetric)

	handler, metriccache := newHandler(storeObjects, externalMetricsListerCache)

	externalItem := getExternalKey(externalMetric)
	err := handler.Process(externalItem)

	if err != nil {
		t.Errorf("error after processing = %v, want %v", err, nil)
	}

	externalRequest, exists := metriccache.GetCloudWatchRequest(externalMetric.Namespace, externalMetric.Name)

	if exists == false {
		t.Errorf("exist = %v, want %v", exists, true)
	}

	validateExternalMetricResult(externalRequest, externalMetric, t)
}

func TestShouldFailOnInvalidCacheKey(t *testing.T) {
	var storeObjects []runtime.Object
	var externalMetricsListerCache []*api.ExternalMetric

	externalMetric := newFullExternalMetric("test")
	storeObjects = append(storeObjects, externalMetric)
	externalMetricsListerCache = append(externalMetricsListerCache, externalMetric)

	handler, metriccache := newHandler(storeObjects, externalMetricsListerCache)

	queueItem := namespacedQueueItem{
		namespaceKey: "invalidkey/with/extrainfo",
		kind:         "somethingwrong",
	}
	err := handler.Process(queueItem)

	if err == nil {
		t.Errorf("error after processing nil, want non nil")
	}

	_, exists := metriccache.GetCloudWatchRequest(externalMetric.Namespace, externalMetric.Name)

	if exists == true {
		t.Errorf("exist = %v, want %v", exists, false)
	}
}

func TestWhenExternalItemHasBeenDeleted(t *testing.T) {
	var storeObjects []runtime.Object
	var externalMetricsListerCache []*api.ExternalMetric

	externalMetric := newFullExternalMetric("test")

	// don't put anything in the stores
	handler, metriccache := newHandler(storeObjects, externalMetricsListerCache)

	// add the item to the cache then test if it gets deleted
	queueItem := getExternalKey(externalMetric)
	metriccache.Update(queueItem.Key(), "test", cloudwatch.GetMetricDataInput{})

	err := handler.Process(queueItem)

	if err != nil {
		t.Errorf("error == %v, want nil", err)
	}

	_, exists := metriccache.GetCloudWatchRequest(externalMetric.Namespace, externalMetric.Name)

	if exists == true {
		t.Errorf("exist = %v, want %v", exists, false)
	}
}

func TestWhenItemKindIsUnknown(t *testing.T) {
	var storeObjects []runtime.Object
	var externalMetricsListerCache []*api.ExternalMetric

	// don't put anything in the stores, as we are not looking anything up
	handler, metriccache := newHandler(storeObjects, externalMetricsListerCache)

	// add the item to the cache then test if it gets deleted
	queueItem := namespacedQueueItem{
		namespaceKey: "default/unknown",
		kind:         "Unknown",
	}

	err := handler.Process(queueItem)

	if err != nil {
		t.Errorf("error == %v, want nil", err)
	}

	_, exists := metriccache.GetCloudWatchRequest("default", "unkown")

	if exists == true {
		t.Errorf("exist = %v, want %v", exists, false)
	}
}

func newHandler(storeObjects []runtime.Object, externalMetricsListerCache []*api.ExternalMetric) (Handler, *metriccache.MetricCache) {
	fakeClient := fake.NewSimpleClientset(storeObjects...)
	i := informers.NewSharedInformerFactory(fakeClient, 0)

	externalMetricLister := i.Metrics().V1alpha1().ExternalMetrics().Lister()

	for _, em := range externalMetricsListerCache {
		i.Metrics().V1alpha1().ExternalMetrics().Informer().GetIndexer().Add(em)
	}

	metriccache := metriccache.NewMetricCache()
	handler := NewHandler(externalMetricLister, metriccache)

	return handler, metriccache
}

func validateExternalMetricResult(metricRequest cloudwatch.GetMetricDataInput, externalMetricInfo *api.ExternalMetric, t *testing.T) {

	// Metric Queries
	if len(metricRequest.MetricDataQueries) != len(externalMetricInfo.Spec.Queries) {
		t.Errorf("metricRequest Queries = %v, want %v", metricRequest.MetricDataQueries, externalMetricInfo.Spec.Queries)
	}

	for i, q := range metricRequest.MetricDataQueries {
		wantQueries := externalMetricInfo.Spec.Queries[i]
		if q.Expression != nil && *q.Expression != wantQueries.Expression {
			t.Errorf("metricRequest Expression = %v, want %v", q.Expression, wantQueries.Expression)
		}

		if *q.Id != wantQueries.ID {
			t.Errorf("metricRequest ID = %v, want %v", q.Id, wantQueries.ID)
		}

		if *q.Label != wantQueries.Label {
			t.Errorf("metricRequest Label = %v, want %v", q.Label, wantQueries.Label)
		}

		qStat := q.MetricStat
		wantStat := wantQueries.MetricStat

		qMetric := qStat.Metric
		wantMetric := wantStat.Metric

		if len(qMetric.Dimensions) != len(wantMetric.Dimensions) {
			t.Errorf("metricRequest Dimensions = %v, want = %v", qMetric.Dimensions, wantMetric.Dimensions)
		}

		for j, d := range qMetric.Dimensions {
			if *d.Name != *qMetric.Dimensions[j].Name {
				t.Errorf("metricRequest Dimension Name = %v, want = %v", d.Name, qMetric.Dimensions[j].Name)
			}

			if *d.Value != *qMetric.Dimensions[j].Value {
				t.Errorf("metricRequest Dimension Value = %v, want = %v", d.Value, qMetric.Dimensions[j].Value)
			}
		}

		if *qMetric.MetricName != wantMetric.MetricName {
			t.Errorf("metricRequest MetricNAme = %v, want %v", qMetric.MetricName, qMetric.MetricName)
		}

		if *qMetric.Namespace != wantMetric.Namespace {
			t.Errorf("metricRequest Namespace = %v, want %v", qMetric.Namespace, qMetric.Namespace)
		}

		if *qStat.Period != wantStat.Period {
			t.Errorf("metricRequest Period = %v, want %v", qStat.Period, wantStat.Period)
		}

		if *qStat.Stat != wantStat.Stat {
			t.Errorf("metricRequest Stat = %v, want %v", qStat.Stat, wantStat.Stat)
		}

		if string(qStat.Unit) != wantStat.Unit {
			t.Errorf("metricRequest Unit = %v, want %v", qStat.Unit, wantStat.Unit)
		}

		if *q.ReturnData != wantQueries.ReturnData {
			t.Errorf("metricRequest ReturnData = %v, want %v", q.ReturnData, wantQueries.ReturnData)
		}
	}

}

func newFullExternalMetric(name string) *api.ExternalMetric {
	// must preserve upper casing for azure api
	return &api.ExternalMetric{
		TypeMeta: metav1.TypeMeta{APIVersion: api.SchemeGroupVersion.String(), Kind: "ExternalMetric"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: api.MetricSeriesSpec{
			Name: "Name",
			Queries: []api.MetricDataQuery{{
				ID: "queryID",
				MetricStat: api.MetricStat{
					Metric: api.Metric{
						Dimensions: []api.Dimension{{
							Name:  "DimensionName",
							Value: "DimensionValue",
						}},
						MetricName: "metricName",
						Namespace:  "namespace",
					},
					Period: 60,
					Stat:   "Average",
					Unit:   "Count",
				},
				ReturnData: true,
			}},
		},
	}
}
