package aws

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/chankh/k8s-cloudwatch-adapter/pkg/config"
)

type Client interface {
	Query(queries []config.MetricDataQuery) ([]cloudwatch.MetricDataResult, error)
}
