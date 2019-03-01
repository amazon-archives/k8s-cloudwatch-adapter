package metriccache

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"

	"github.com/golang/glog"
)

// MetricCache holds the loaded metric request info in the system
type MetricCache struct {
	metricMutext   sync.RWMutex
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
	mc.metricMutext.Lock()
	defer mc.metricMutext.Unlock()

	mc.metricRequests[key] = metricRequest
	mc.metricNames[key] = name
}

// GetAzureMonitorRequest retrieves a metric request from the cache
func (mc *MetricCache) GetCloudWatchRequest(namepace, name string) (cloudwatch.GetMetricDataInput, bool) {
	mc.metricMutext.RLock()
	defer mc.metricMutext.RUnlock()

	key := externalMetricKey(namepace, name)
	metricRequest, exists := mc.metricRequests[key]
	if !exists {
		glog.V(2).Infof("metric not found %s", key)
		return cloudwatch.GetMetricDataInput{}, false
	}

	return metricRequest.(cloudwatch.GetMetricDataInput), true
}

// Remove retrieves a metric request from the cache
func (mc *MetricCache) Remove(key string) {
	mc.metricMutext.Lock()
	defer mc.metricMutext.Unlock()

	delete(mc.metricRequests, key)
	delete(mc.metricNames, key)
}

func (mc *MetricCache) Keys() []string {
	keys := make([]string, len(mc.metricNames))
	for k := range mc.metricNames {
		keys = append(keys, mc.metricNames[k])
	}

	return keys
}

func externalMetricKey(namespace string, name string) string {
	return fmt.Sprintf("ExternalMetric/%s/%s", namespace, name)
}
