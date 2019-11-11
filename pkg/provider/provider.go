package provider

import (
	"sync"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/aws"
	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/metriccache"
	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
)

var nsGroupResource = schema.GroupResource{Resource: "namespaces"}

// cloudwatchProvider is a implementation of provider.MetricsProvider for CloudWatch
type cloudwatchProvider struct {
	client   dynamic.Interface
	mapper   apimeta.RESTMapper
	cwClient aws.Client

	valuesLock  sync.RWMutex
	metricCache *metriccache.MetricCache
}

// NewCloudWatchProvider returns an instance of testingProvider, along with its restful.WebService
// that opens endpoints to post new fake metrics
func NewCloudWatchProvider(client dynamic.Interface, mapper apimeta.RESTMapper, cwClient aws.Client, metricCache *metriccache.MetricCache) provider.ExternalMetricsProvider {
	provider := &cloudwatchProvider{
		client:      client,
		mapper:      mapper,
		cwClient:    cwClient,
		metricCache: metricCache,
	}

	return provider
}
