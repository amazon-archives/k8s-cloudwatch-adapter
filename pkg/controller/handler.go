package controller

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	listers "github.com/awslabs/k8s-cloudwatch-adapter/pkg/client/listers/metrics/v1alpha1"
	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/metriccache"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
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
	klog.V(2).Infof("processing item '%s' in namespace '%s'", name, ns)
	externalMetricInfo, err := h.externalmetricLister.ExternalMetrics(ns).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			// Then this we should remove
			klog.V(2).Infof("removing item from cache '%s' in namespace '%s'", name, ns)
			h.metriccache.Remove(queueItem.Key())
			return nil
		}

		return err
	}

	klog.V(2).Infof("externalMetricInfo: %v", externalMetricInfo)
	queries := externalMetricInfo.Spec.Queries

	// If changing logic in this block ensure changes are duplicated in
	// `pkg/client.Query()`
	cwMetricQueries := make([]*cloudwatch.MetricDataQuery, len(queries))
	for i, q := range queries {
		q := q
		mdq := &cloudwatch.MetricDataQuery{
			Id:         &q.ID,
			Label:      &q.Label,
			ReturnData: &q.ReturnData,
		}

		if len(q.Expression) == 0 {
			dimensions := make([]*cloudwatch.Dimension, len(q.MetricStat.Metric.Dimensions))
			for j := range q.MetricStat.Metric.Dimensions {
				dimensions[j] = &cloudwatch.Dimension{
					Name:  &q.MetricStat.Metric.Dimensions[j].Name,
					Value: &q.MetricStat.Metric.Dimensions[j].Value,
				}
			}

			metric := &cloudwatch.Metric{
				Dimensions: dimensions,
				MetricName: &q.MetricStat.Metric.MetricName,
				Namespace:  &q.MetricStat.Metric.Namespace,
			}

			mdq.MetricStat = &cloudwatch.MetricStat{
				Metric: metric,
				Period: &q.MetricStat.Period,
				Stat:   &q.MetricStat.Stat,
				Unit:   aws.String(q.MetricStat.Unit),
			}
		} else {
			mdq.Expression = &q.Expression
		}

		cwMetricQueries[i] = mdq
	}
	cwQuery := cloudwatch.GetMetricDataInput{
		MetricDataQueries: cwMetricQueries,
	}

	klog.V(2).Infof("adding to cache item '%s' in namespace '%s'", name, ns)
	h.metriccache.Update(queueItem.Key(), name, cwQuery)

	return nil
}
