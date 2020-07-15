package aws

import (
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// Client represents a client for Amazon CloudWatch.
type Client interface {

	// Query sends a CloudWatch GetMetricDataInput to CloudWatch API for metric results.
	QueryCloudWatch(query cloudwatch.GetMetricDataInput) ([]*cloudwatch.MetricDataResult, error)
}
