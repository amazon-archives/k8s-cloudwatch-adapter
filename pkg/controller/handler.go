package controller

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	listers "github.com/awslabs/k8s-cloudwatch-adapter/pkg/client/listers/metrics/v1alpha1"
	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/metriccache"
	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// Handler processes the events from the controler for external metrics
type Handler struct {
	externalmetricLister listers.ExternalMetricLister
	metriccache          *metriccache.MetricCache
}

// NewHandler created a new handler
func NewHandler(externalmetricLister listers.ExternalMetricLister, metricCache *metriccache.MetricCache) Handler {
	return Handler{
		externalmetricLister: externalmetricLister,
		metriccache:          metricCache,
	}
}

// ControllerHandler is a handler to process resource items
type ControllerHandler interface {
	Process(queueItem namespacedQueueItem) error
}

// Process validates the item exists then stores updates the metric cached used to make requests to
// cloudwatch
func (h *Handler) Process(queueItem namespacedQueueItem) error {
	ns, name, err := cache.SplitMetaNamespaceKey(queueItem.namespaceKey)
	if err != nil {
		// not a valid key do not put back on queue
		runtime.HandleError(fmt.Errorf("expected namespace/name key in workqueue but got %s", queueItem.namespaceKey))
		return err
	}

	switch queueItem.kind {
	case "ExternalMetric":
		return h.handleExternalMetric(ns, name, queueItem)
	}

	return nil
}

func (h *Handler) handleExternalMetric(ns, name string, queueItem namespacedQueueItem) error {
	// check if item exists
	glog.V(2).Infof("processing item '%s' in namespace '%s'", name, ns)
	externalMetricInfo, err := h.externalmetricLister.ExternalMetrics(ns).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			// Then this we should remove
			glog.V(2).Infof("removing item from cache '%s' in namespace '%s'", name, ns)
			h.metriccache.Remove(queueItem.Key())
			return nil
		}

		return err
	}

	glog.V(2).Infof("externalMetricInfo: %v", externalMetricInfo)
	queries := externalMetricInfo.Spec.Queries
	cwMetricQueries := make([]cloudwatch.MetricDataQuery, len(queries))
	for i, q := range queries {
		dimensions := make([]cloudwatch.Dimension, len(q.MetricStat.Metric.Dimensions))
		for j, d := range q.MetricStat.Metric.Dimensions {
			dimensions[j] = cloudwatch.Dimension{
				Name:  &d.Name,
				Value: &d.Value,
			}
		}
		metric := &cloudwatch.Metric{
			Dimensions: dimensions,
			MetricName: &q.MetricStat.Metric.MetricName,
			Namespace:  &q.MetricStat.Metric.Namespace,
		}
		unit := cloudwatch.StandardUnit(q.MetricStat.Unit)
		var metricStat *cloudwatch.MetricStat
		expression := q.Expression
		id := q.ID
		var e *string

		if len(expression) == 0 {
			e = nil
			metricStat = &cloudwatch.MetricStat{
				Metric: metric,
				Period: &q.MetricStat.Period,
				Stat:   &q.MetricStat.Stat,
				Unit:   unit,
			}
		} else {
			e = &expression
			metricStat = nil
		}
		cwMetricQueries[i] = cloudwatch.MetricDataQuery{
			Expression: e,
			Id:         &id,
			Label:      &q.Label,
			MetricStat: metricStat,
			ReturnData: &q.ReturnData,
		}
	}

	cwMetricRequest := cloudwatch.GetMetricDataInput{
		MetricDataQueries: cwMetricQueries,
	}

	glog.V(2).Infof("adding to cache item '%s' in namespace '%s'", name, ns)
	h.metriccache.Update(queueItem.Key(), name, cwMetricRequest)

	return nil
}
