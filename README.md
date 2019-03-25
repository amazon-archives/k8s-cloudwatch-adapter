[![Build Status](https://travis-ci.org/awslabs/k8s-cloudwatch-adapter.svg?branch=master)](https://travis-ci.org/awslabs/k8s-cloudwatch-adapter)
[![GitHub
release](https://img.shields.io/github/release/awslabs/k8s-cloudwatch-adapter/all.svg)](https://github.com/awslabs/k8s-cloudwatch-adapter/releases)
[![docker image
size](https://shields.beevelop.com/docker/image/image-size/chankh/k8s-cloudwatch-adapter/latest.svg)](https://hub.docker.com/r/chankh/k8s-cloudwatch-adapter)
[![image
layers](https://shields.beevelop.com/docker/image/layers/chankh/k8s-cloudwatch-adapter/latest.svg)](https://hub.docker.com/r/chankh/k8s-cloudwatch-adapter)
[![image
pulls](https://shields.beevelop.com/docker/pulls/chankh/k8s-cloudwatch-adapter.svg)](https://hub.docker.com/r/chankh/k8s-cloudwatch-adapter)

# Kubernetes Custom Metrics Adapter for Kubernetes


An implementation of the Kubernetes [Custom Metrics API and External Metrics
API](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-metrics-apis)
for AWS CloudWatch metrics.

This adapter allows you to scale your Kubernetes deployment using the [Horizontal Pod
Autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) (HPA) with
metrics from AWS CloudWatch.

**This project is currently in Alpha status, use at your own risk.**

## Prerequsites
This adapter requires the following permissions to access metric data from Amazon CloudWatch.
- cloudwatch:GetMetricData
- cloudwatch:GetMetricStatistics
- cloudwatch:ListMetrics

You can create an IAM policy using this template, and attach it to the role if you are using
[kube2iam](https://github.com/jtblin/kube2iam). 

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "cloudwatch:GetMetricData",
                "cloudwatch:GetMetricStatistics",
                "cloudwatch:ListMetrics"
            ],
            "Resource": "*"
        }
    ]
}
```

## Deploy
Requires a Kubernetes cluster with Metric Server deployed, Amazon EKS cluster is fine too.

Now deploy the adapter to your Kubernetes cluster.

```bash
$ kubectl apply -f https://raw.githubusercontent.com/awslabs/k8s-cloudwatch-adapter/master/deploy/adapter.yaml
namespace/custom-metrics created
clusterrolebinding.rbac.authorization.k8s.io/k8s-cloudwatch-adapter:system:auth-delegator created
rolebinding.rbac.authorization.k8s.io/k8s-cloudwatch-adapter-auth-reader created
deployment.apps/k8s-cloudwatch-adapter configured
clusterrolebinding.rbac.authorization.k8s.io/k8s-cloudwatch-adapter-resource-reader created
serviceaccount/k8s-cloudwatch-adapter created
service/k8s-cloudwatch-adapter created
apiservice.apiregistration.k8s.io/v1beta1.custom.metrics.k8s.io created
apiservice.apiregistration.k8s.io/v1beta1.external.metrics.k8s.io created
clusterrole.rbac.authorization.k8s.io/k8s-cloudwatch-adapter:custom-metrics-reader created
clusterrole.rbac.authorization.k8s.io/k8s-cloudwatch-adapter:external-metrics-reader created
clusterrole.rbac.authorization.k8s.io/k8s-cloudwatch-adapter-resource-reader created
clusterrolebinding.rbac.authorization.k8s.io/k8s-cloudwatch-adapter:custom-metrics-reader created
clusterrolebinding.rbac.authorization.k8s.io/k8s-cloudwatch-adapter:external-metrics-reader created
customresourcedefinition.apiextensions.k8s.io/externalmetrics.metrics.aws created
clusterrole.rbac.authorization.k8s.io/k8s-cloudwatch-adapter:crd-metrics-reader created
clusterrolebinding.rbac.authorization.k8s.io/k8s-cloudwatch-adapter:crd-metrics-reader created
```

This creates a new namespace `custom-metrics` and deploys the necessary ClusterRole, Service Account,
Role Binding, along with the deployment of the adapter.

### Verifying the deployment
Next you can query the APIs to see if the adapter is deployed correctly by running:

```bash
$ kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1" | jq .
{
  "kind": "APIResourceList",
  "apiVersion": "v1",
  "groupVersion": "custom.metrics.k8s.io/v1beta1",
  "resources": [
  ]
}
```

and 

```bash
$ kubectl get --raw "/apis/external.metrics.k8s.io/v1beta1" | jq.
{
  "kind": "APIResourceList",
  "apiVersion": "v1",
  "groupVersion": "external.metrics.k8s.io/v1beta1",
  "resources": [
  ]
}
```

## Deploying the sample application
There is a sample SQS application provided in this repository for you to test how the adapter works.
Refer to this [guide](samples/sqs/README.md)

## License

This library is licensed under the Apache 2.0 License. 

## Issues
Report any issues in the [Github Issues](https://github.com/awslabs/k8s-cloudwatch-adapter/issues)
