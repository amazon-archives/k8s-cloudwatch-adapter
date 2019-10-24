package aws

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/config"
	"github.com/golang/glog"
)

// NewCloudWatchClient creates a new CloudWatch client.
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
	glog.Infof("using AWS Region: %s", cfg.Region)

	if os.Getenv("DEBUG") == "true" {
		cfg.LogLevel = aws.LogDebugWithHTTPBody
	}

	// Using the Config value, create the CloudWatch client
	svc := cloudwatch.New(cfg)
	return &cloudwatchClient{client: svc}
}

type cloudwatchClient struct {
	client *cloudwatch.CloudWatch
}

func (c *cloudwatchClient) Query(queries []config.MetricDataQuery) ([]cloudwatch.MetricDataResult, error) {
	// If changing logic in this function ensure changes are duplicated in
	// `pkg/controller.handleExternalMetric()`
	cwMetricQueries := make([]cloudwatch.MetricDataQuery, len(queries))
	for i, q := range queries {
		q := q
		mdq := cloudwatch.MetricDataQuery{
			Id:         &q.ID,
			Label:      &q.Label,
			ReturnData: &q.ReturnData,
		}

		if len(q.Expression) == 0 {
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

			mdq.MetricStat = &cloudwatch.MetricStat{
				Metric: metric,
				Period: &q.MetricStat.Period,
				Stat:   &q.MetricStat.Stat,
				Unit:   cloudwatch.StandardUnit(q.MetricStat.Unit),
			}
		} else {
			mdq.Expression = &q.Expression
		}

		cwMetricQueries[i] = mdq
	}
	cwQuery := cloudwatch.GetMetricDataInput{
		MetricDataQueries: cwMetricQueries,
	}

	return c.QueryCloudWatch(cwQuery)
}

func (c *cloudwatchClient) QueryCloudWatch(cwQuery cloudwatch.GetMetricDataInput) ([]cloudwatch.MetricDataResult, error) {
	now := time.Now()
	endTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
	// CloudWatch metrics have latency, we will grab in a 5 minute window and extract the latest value
	startTime := endTime.Add(-5 * time.Minute)

	cwQuery.EndTime = &endTime
	cwQuery.StartTime = &startTime
	cwQuery.ScanBy = "TimestampDescending"

	results, err := c.client.GetMetricDataRequest(&cwQuery).Send(context.Background())
	if err != nil {
		glog.Errorf("err: %v", err)
		return []cloudwatch.MetricDataResult{}, err
	}

	return results.MetricDataResults, nil
}
