package aws

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/config"
	"k8s.io/klog"
)

// NewCloudWatchClient creates a new CloudWatch client.
func NewCloudWatchClient() Client {
	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg := aws.NewConfig()

	// check if region is set
	if aws.StringValue(cfg.Region) == "" {
		cfg.Region = aws.String(GetLocalRegion())
	}
	klog.Infof("using AWS Region: %s", aws.StringValue(cfg.Region))

	if os.Getenv("DEBUG") == "true" {
		cfg = cfg.WithLogLevel(aws.LogDebugWithHTTPBody)
	}

	// Using the Config value, create the CloudWatch client
	sess := session.Must(session.NewSession(cfg))
	svc := cloudwatch.New(sess)
	return &cloudwatchClient{client: svc}
}

type cloudwatchClient struct {
	client *cloudwatch.CloudWatch
}

func (c *cloudwatchClient) Query(queries []config.MetricDataQuery) ([]*cloudwatch.MetricDataResult, error) {
	// If changing logic in this function ensure changes are duplicated in
	// `pkg/controller.handleExternalMetric()`
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
			for j, d := range q.MetricStat.Metric.Dimensions {
				dimensions[j] = &cloudwatch.Dimension{
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
				Unit:   &q.MetricStat.Unit,
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

func (c *cloudwatchClient) QueryCloudWatch(cwQuery cloudwatch.GetMetricDataInput) ([]*cloudwatch.MetricDataResult, error) {
	now := time.Now()
	endTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
	// CloudWatch metrics have latency, we will grab in a 5 minute window and extract the latest value
	startTime := endTime.Add(-5 * time.Minute)

	cwQuery.EndTime = &endTime
	cwQuery.StartTime = &startTime
	cwQuery.ScanBy = aws.String("TimestampDescending")

	req, resp := c.client.GetMetricDataRequest(&cwQuery)
	req.SetContext(context.Background())

	if err := req.Send(); err != nil {
		klog.Errorf("err: %v", err)
		return []*cloudwatch.MetricDataResult{}, err
	}

	return resp.MetricDataResults, nil
}
