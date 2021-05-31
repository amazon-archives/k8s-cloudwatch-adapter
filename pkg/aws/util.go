package aws

import (
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/apis/metrics/v1alpha1"

	"k8s.io/klog"
)

// GetLocalRegion gets the region ID from the instance metadata.
func GetLocalRegion() string {
	resp, err := http.Get("http://169.254.169.254/latest/meta-data/placement/availability-zone/")
	if err != nil {
		klog.Errorf("unable to get current region information, %v", err)
		return ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		klog.Errorf("cannot read response from instance metadata, %v", err)
	}

	// strip the last character from AZ to get region ID
	return string(body[0 : len(body)-1])
}

func toCloudWatchQuery(externalMetric *v1alpha1.ExternalMetric) cloudwatch.GetMetricDataInput {
	queries := externalMetric.Spec.Queries

	cwMetricQueries := make([]types.MetricDataQuery, len(queries))
	for i, q := range queries {
		q := q
		returnData := &q.ReturnData
		mdq := types.MetricDataQuery{
			Id:         &q.ID,
			Label:      &q.Label,
			ReturnData: *returnData,
		}

		if len(q.Expression) == 0 {
			dimensions := make([]types.Dimension, len(q.MetricStat.Metric.Dimensions))
			for j := range q.MetricStat.Metric.Dimensions {
				dimensions[j] = types.Dimension{
					Name:  &q.MetricStat.Metric.Dimensions[j].Name,
					Value: &q.MetricStat.Metric.Dimensions[j].Value,
				}
			}

			metric := &types.Metric{
				Dimensions: dimensions,
				MetricName: &q.MetricStat.Metric.MetricName,
				Namespace:  &q.MetricStat.Metric.Namespace,
			}

			mdq.MetricStat = &types.MetricStat{
				Metric: metric,
				Period: &q.MetricStat.Period,
				Stat:   &q.MetricStat.Stat,
				Unit:   types.StandardUnit(q.MetricStat.Unit),
			}
		} else {
			mdq.Expression = &q.Expression
		}

		cwMetricQueries[i] = mdq
	}
	cwQuery := cloudwatch.GetMetricDataInput{
		MetricDataQueries: cwMetricQueries,
	}

	return cwQuery
}
