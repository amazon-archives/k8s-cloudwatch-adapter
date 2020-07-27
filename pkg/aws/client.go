package aws

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"

	"github.com/aws/aws-sdk-go/aws/endpoints"

	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/apis/metrics/v1alpha1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"k8s.io/klog"
)

func NewCloudWatchManager() CloudWatchManager {
	return &cloudwatchManager{}
}

type cloudwatchManager struct {
}

func (c *cloudwatchManager) getClient(role, region string) *cloudwatch.CloudWatch {
	// Using the Config value, create the CloudWatch client
	sess := session.Must(session.NewSession())

	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg := aws.NewConfig().WithSTSRegionalEndpoint(endpoints.RegionalSTSEndpoint)

	// check if roleARN is passed
	if role != "" {
		creds := stscreds.NewCredentials(sess, role)
		cfg = cfg.WithCredentials(creds)
		klog.Infof("using IAM role ARN: %s", role)
	}

	// check if region is set
	if region != "" {
		cfg = cfg.WithRegion(region)
	} else if aws.StringValue(cfg.Region) == "" {
		cfg.Region = aws.String(GetLocalRegion())
	}
	klog.Infof("using AWS Region: %s", aws.StringValue(cfg.Region))

	if os.Getenv("DEBUG") == "true" {
		cfg = cfg.WithLogLevel(aws.LogDebugWithHTTPBody)
	}

	svc := cloudwatch.New(sess, cfg)
	return svc
}

func (c *cloudwatchManager) QueryCloudWatch(request v1alpha1.ExternalMetric) ([]*cloudwatch.MetricDataResult, error) {
	role := request.Spec.RoleARN
	region := request.Spec.Region
	cwQuery := toCloudWatchQuery(&request)
	now := time.Now()
	endTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
	// CloudWatch metrics have latency, we will grab in a 5 minute window and extract the latest value
	startTime := endTime.Add(-5 * time.Minute)

	cwQuery.EndTime = &endTime
	cwQuery.StartTime = &startTime
	cwQuery.ScanBy = aws.String("TimestampDescending")

	req, resp := c.getClient(role, region).GetMetricDataRequest(&cwQuery)
	req.SetContext(context.Background())

	if err := req.Send(); err != nil {
		klog.Errorf("err: %v", err)
		return []*cloudwatch.MetricDataResult{}, err
	}

	return resp.MetricDataResults, nil
}
