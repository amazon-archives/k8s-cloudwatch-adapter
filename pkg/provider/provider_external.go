package provider

import (
	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/metrics/pkg/apis/external_metrics"
)

func (p *cloudwatchProvider) GetExternalMetric(namespace string, metricSelector labels.Selector, info provider.ExternalMetricInfo) (*external_metrics.ExternalMetricValueList, error) {
	p.valuesLock.RLock()
	defer p.valuesLock.RUnlock()

	externalMetric := external_metrics.ExternalMetricValue{
		MetricName: info.Metric,
		Value:      *resource.NewQuantity(int64(0), resource.DecimalSI),
		Timestamp:  metav1.Now(),
	}
	matchingMetrics := []external_metrics.ExternalMetricValue{}
	matchingMetrics = append(matchingMetrics, externalMetric)

	return &external_metrics.ExternalMetricValueList{
		Items: matchingMetrics,
	}, nil
}

func (p *cloudwatchProvider) ListAllExternalMetrics() []provider.ExternalMetricInfo {
	p.valuesLock.RLock()
	defer p.valuesLock.RUnlock()

	// not implemented yet
	externalMetricsInfo := []provider.ExternalMetricInfo{}
	return externalMetricsInfo
}
