package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/apis/metrics/v1alpha1"
)

// CloudWatchManager manages clients for Amazon CloudWatch.
type CloudWatchManager interface {
	// Query sends a CloudWatch GetMetricDataInput to CloudWatch API for metric results.
	QueryCloudWatch(request v1alpha1.ExternalMetric) ([]types.MetricDataResult, error)
}
