[![Build Status](https://travis-ci.org/awslabs/k8s-cloudwatch-adapter.svg?branch=master)](https://travis-ci.org/awslabs/k8s-cloudwatch-adapter)
[![GitHub
release](https://img.shields.io/github/release/awslabs/k8s-cloudwatch-adapter/all.svg)](https://github.com/awslabs/k8s-cloudwatch-adapter/releases)
[![docker image
size](https://shields.beevelop.com/docker/image/image-size/chankh/k8s-cloudwatch-adapter/latest.svg)](https://hub.docker.com/r/chankh/k8s-cloudwatch-adapter)
[![image
layers](https://shields.beevelop.com/docker/image/layers/chankh/k8s-cloudwatch-adapter/latest.svg)](https://hub.docker.com/r/chankh/k8s-cloudwatch-adapter)
[![image
pulls](https://shields.beevelop.com/docker/pulls/chankh/k8s-cloudwatch-adapter.svg)](https://hub.docker.com/r/chankh/k8s-cloudwatch-adapter)

> Attention! This project has been archived and is no longer being worked on. If you are looking for a metrics server that can consume metrics from CloudWatch, please consider using the [KEDA](https://keda.sh) project instead. KEDA is a Kubernetes-based Event Driven Autoscaler. With KEDA, you can drive the scaling of any container in Kubernetes based on the number of events needing to be processed. For an overview of KEDA, see [An overview of Kubernetes Event-Driven Autoscaling](https://youtu.be/H5eZEq_wqSE).

# Kubernetes Custom Metrics Adapter for Kubernetes


An implementation of the Kubernetes [Custom Metrics API and External Metrics
API](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-metrics-apis)
for AWS CloudWatch metrics.

This adapter allows you to scale your Kubernetes deployment using the [Horizontal Pod
Autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) (HPA) with
metrics from AWS CloudWatch.

## Prerequisites
This adapter requires the following permissions to access metric data from Amazon CloudWatch.
- cloudwatch:GetMetricData

You can create an IAM policy using this template, and attach it to the [Service Account Role](https://docs.aws.amazon.com/eks/latest/userguide/specify-service-account-role.html) if you are using
[IAM Roles for Service Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html).

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "cloudwatch:GetMetricData"
            ],
            "Resource": "*"
        }
    ]
}
```

## Deploy
Requires a Kubernetes cluster with Metric Server deployed, Amazon EKS cluster is fine too.

Now deploy the adapter to your Kubernetes cluster:

```bash
$ kubectl apply -f https://raw.githubusercontent.com/awslabs/k8s-cloudwatch-adapter/master/deploy/adapter.yaml
namespace/custom-metrics created
clusterrolebinding.rbac.authorization.k8s.io/k8s-cloudwatch-adapter:system:auth-delegator created
rolebinding.rbac.authorization.k8s.io/k8s-cloudwatch-adapter-auth-reader created
deployment.apps/k8s-cloudwatch-adapter created
clusterrolebinding.rbac.authorization.k8s.io/k8s-cloudwatch-adapter-resource-reader created
serviceaccount/k8s-cloudwatch-adapter created
service/k8s-cloudwatch-adapter created
apiservice.apiregistration.k8s.io/v1beta1.external.metrics.k8s.io created
clusterrole.rbac.authorization.k8s.io/k8s-cloudwatch-adapter:external-metrics-reader created
clusterrole.rbac.authorization.k8s.io/k8s-cloudwatch-adapter-resource-reader created
clusterrolebinding.rbac.authorization.k8s.io/k8s-cloudwatch-adapter:external-metrics-reader created
customresourcedefinition.apiextensions.k8s.io/externalmetrics.metrics.aws created
clusterrole.rbac.authorization.k8s.io/k8s-cloudwatch-adapter:crd-metrics-reader created
clusterrolebinding.rbac.authorization.k8s.io/k8s-cloudwatch-adapter:crd-metrics-reader created
```

This creates a new namespace `custom-metrics` and deploys the necessary ClusterRole, Service Account,
Role Binding, along with the deployment of the adapter.

Alternatively the crd and adapter can be deployed using the Helm chart in the `/charts` directory:

```bash
$ helm install k8s-cloudwatch-adapter-crd ./charts/k8s-cloudwatch-adapter-crd
NAME: k8s-cloudwatch-adapter-crd
LAST DEPLOYED: Thu Sep 17 11:36:53 2020
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None
$ helm install k8s-cloudwatch-adapter ./charts/k8s-cloudwatch-adapter \
>   --namespace custom-metrics \
>   --create-namespace
NAME: k8s-cloudwatch-adapter
LAST DEPLOYED: Fri Aug 14 13:20:17 2020
NAMESPACE: custom-metrics
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

### Verifying the deployment
Next you can query the APIs to see if the adapter is deployed correctly by running:

```bash
$ kubectl get --raw "/apis/external.metrics.k8s.io/v1beta1" | jq .
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
Refer to this [guide](samples/sqs/README.md).

## More docs
- [Configuring cross account metric example](docs/cross-account.md)
- [ExternalMetric CRD schema](docs/schema.md)

## License

This library is licensed under the Apache 2.0 License.

## Issues
Report any issues in the [Github Issues](https://github.com/awslabs/k8s-cloudwatch-adapter/issues)
