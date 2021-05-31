package aws

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/apis/metrics/v1alpha1"

	"k8s.io/klog"
)

func NewCloudWatchManager() CloudWatchManager {
	return &cloudwatchManager{
		localRegion: GetLocalRegion(),
	}
}

type cloudwatchManager struct {
	localRegion string
}

func (c *cloudwatchManager) getClient(role, region *string) *cloudwatch.Client {
	// check if region is set
	usedRegion := c.localRegion
	if region != nil {
		usedRegion = *region
	}

	// Using the SDK's default configuration, loading additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(usedRegion))
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	klog.Infof("using AWS Region: %s", cfg.Region)

	// check if roleARN is passed
	if role != nil {
		client := sts.NewFromConfig(cfg)
		provider := stscreds.NewAssumeRoleProvider(client, *role)
		cfg.Credentials = aws.NewCredentialsCache(provider)
		klog.Infof("using IAM role ARN: %s", *role)
	}

	if os.Getenv("DEBUG") == "true" {
		cfg.ClientLogMode = aws.LogRequestWithBody | aws.LogResponseWithBody
	}

	client := cloudwatch.NewFromConfig(cfg)
	return client
}

func (c *cloudwatchManager) QueryCloudWatch(request v1alpha1.ExternalMetric) ([]types.MetricDataResult, error) {
	role := request.Spec.RoleARN
	region := request.Spec.Region
	cwQuery := toCloudWatchQuery(&request)
	now := time.Now()
	endTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
	// CloudWatch metrics have latency, we will grab in a 5 minute window and extract the latest value
	startTime := endTime.Add(-5 * time.Minute)

	cwQuery.EndTime = &endTime
	cwQuery.StartTime = &startTime
	cwQuery.ScanBy = types.ScanByTimestampDescending

	getMetricDataOutput, err := c.getClient(role, region).GetMetricData(context.Background(), &cwQuery)

	if err != nil {
		klog.Errorf("err: %v", err)
		return []types.MetricDataResult{}, err
	}

	return getMetricDataOutput.MetricDataResults, nil
}
