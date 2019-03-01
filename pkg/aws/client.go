package aws

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/chankh/k8s-cloudwatch-adapter/pkg/config"
	"github.com/golang/glog"
)

func NewCloudWatchClient() Client {
	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	// check if region is set
	if cfg.Region == "" {
		cfg.Region = GetLocalRegion()
	}

	// Using the Config value, create the CloudWatch client
	svc := cloudwatch.New(cfg)
	return &cloudwatchClient{client: svc}
}

type cloudwatchClient struct {
	client *cloudwatch.CloudWatch
}

func (c *cloudwatchClient) Query(queries []config.MetricDataQuery) ([]cloudwatch.MetricDataResult, error) {
	now := time.Now()
	endTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
	// CloudWatch metrics have latency, we will grab in a 5 minute window and extract the latest value
	startTime := endTime.Add(-5 * time.Minute)

	mdq := make([]cloudwatch.MetricDataQuery, len(queries))
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
		mdq[i] = cloudwatch.MetricDataQuery{
			Expression: e,
			Id:         &id,
			Label:      &q.Label,
			MetricStat: metricStat,
			ReturnData: &q.ReturnData,
		}
	}

	cwQuery := &cloudwatch.GetMetricDataInput{
		EndTime:           &endTime,
		MetricDataQueries: mdq,
		ScanBy:            "TimestampDescending",
		StartTime:         &startTime,
	}

	results, err := c.client.GetMetricDataRequest(cwQuery).Send()
	if err != nil {
		glog.Errorf("err: %v", err)
		return []cloudwatch.MetricDataResult{}, err
	}

	return results.MetricDataResults, nil
}
