package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/chankh/k8s-cloudwatch-adapter/pkg/config"
)

// Client represents a client for Amazon CloudWatch.
type Client interface {
	// Query sends a list of queries to Cloudwatch for metric results.
	Query(queries []config.MetricDataQuery) ([]cloudwatch.MetricDataResult, error)

	// Query sends a CloudWatch GetMetricDataInput to CloudWatch API for metric results.
	QueryCloudWatch(query cloudwatch.GetMetricDataInput) ([]cloudwatch.MetricDataResult, error)
}
