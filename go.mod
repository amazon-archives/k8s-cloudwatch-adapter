module github.com/awslabs/k8s-cloudwatch-adapter

go 1.15

require (
	github.com/aws/aws-sdk-go-v2 v1.6.0
	github.com/aws/aws-sdk-go-v2/config v1.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.2.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sqs v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.4.1 // indirect
	github.com/go-sql-driver/mysql v1.4.0 // indirect
	github.com/kubernetes-incubator/custom-metrics-apiserver v0.0.0-20200323093244-5046ce1afe6b
	github.com/pkg/errors v0.9.1
	gopkg.in/yaml.v2 v2.2.8 // indirect
	k8s.io/apimachinery v0.17.7
	k8s.io/apiserver v0.17.7 // indirect
	k8s.io/client-go v0.17.7
	k8s.io/code-generator v0.17.7
	k8s.io/component-base v0.17.7
	k8s.io/klog v1.0.0
	k8s.io/metrics v0.17.7
)
