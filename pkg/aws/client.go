package aws

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
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
