package metriccache

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"

	"github.com/golang/glog"
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

// GetCloudWatchRequest retrieves a metric request from the cache
func (mc *MetricCache) GetCloudWatchRequest(namepace, name string) (cloudwatch.GetMetricDataInput, bool) {
	mc.metricMutex.RLock()
	defer mc.metricMutex.RUnlock()

	key := externalMetricKey(namepace, name)
	metricRequest, exists := mc.metricRequests[key]
	if !exists {
		glog.V(2).Infof("metric not found %s", key)
		return cloudwatch.GetMetricDataInput{}, false
	}

	return metricRequest.(cloudwatch.GetMetricDataInput), true
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
