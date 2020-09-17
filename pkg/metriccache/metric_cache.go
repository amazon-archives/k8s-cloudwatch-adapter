package metriccache

import (
	"fmt"
	"sync"

	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/apis/metrics/v1alpha1"

	"k8s.io/klog"
)

// MetricCache holds the loaded metric request info in the system
type MetricCache struct {
	metricMutex    sync.RWMutex
	metricRequests map[string]interface{}
	metricNames    map[string]string
}

// NewMetricCache creates the cache
func NewMetricCache() *MetricCache {
	return &MetricCache{
		metricRequests: make(map[string]interface{}),
		metricNames:    make(map[string]string),
	}
}

// Update sets a metric request in the cache
func (mc *MetricCache) Update(key string, name string, metricRequest interface{}) {
	mc.metricMutex.Lock()
	defer mc.metricMutex.Unlock()

	mc.metricRequests[key] = metricRequest
	mc.metricNames[key] = name
}

// GetExternalMetric retrieves an external metric request from the cache
func (mc *MetricCache) GetExternalMetric(namespace, name string) (v1alpha1.ExternalMetric, bool) {
	mc.metricMutex.RLock()
	defer mc.metricMutex.RUnlock()

	key := externalMetricKey(namespace, name)
	metricRequest, exists := mc.metricRequests[key]
	if !exists {
		klog.V(2).Infof("metric not found %s", key)
		return v1alpha1.ExternalMetric{}, false
	}

	return metricRequest.(v1alpha1.ExternalMetric), true
}

// Remove removes a metric request from the cache
func (mc *MetricCache) Remove(key string) {
	mc.metricMutex.Lock()
	defer mc.metricMutex.Unlock()

	delete(mc.metricRequests, key)
	delete(mc.metricNames, key)
}

// ListMetricNames retrieves a list of metric names from the cache.
func (mc *MetricCache) ListMetricNames() []string {
	keys := make([]string, len(mc.metricNames))
	for k := range mc.metricNames {
		keys = append(keys, mc.metricNames[k])
	}

	return keys
}

func externalMetricKey(namespace string, name string) string {
	return fmt.Sprintf("ExternalMetric/%s/%s", namespace, name)
}
