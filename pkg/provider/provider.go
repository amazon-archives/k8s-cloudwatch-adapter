package provider

import (
	"sync"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/dynamic"

	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/aws"
	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/metriccache"
	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
)

// cloudwatchProvider is a implementation of provider.MetricsProvider for CloudWatch
type cloudwatchProvider struct {
	client    dynamic.Interface
	mapper    apimeta.RESTMapper
	cwManager aws.CloudWatchManager

	valuesLock  sync.RWMutex
	metricCache *metriccache.MetricCache
}

// NewCloudWatchProvider returns an instance of cloudwatchProvider
func NewCloudWatchProvider(client dynamic.Interface, mapper apimeta.RESTMapper, cwManager aws.CloudWatchManager, metricCache *metriccache.MetricCache) provider.ExternalMetricsProvider {
	provider := &cloudwatchProvider{
		client:      client,
		mapper:      mapper,
		cwManager:   cwManager,
		metricCache: metricCache,
	}

	return provider
}
