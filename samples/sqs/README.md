# Sample Amazon SQS Application

This directory contains a sample Amazon SQS application to test out `k8s-cloudwatch-adapter`. SQS
producer and consumer are provided, together with the YAML files for deploying the consumer, metric
configuration and HPA.

Both the producer and consumer will use an Amazon SQS queue named `helloworld`. This queue will be
created by the producer if it does not exist.

## Prerequisites

Before starting, you need to first grant permissions to your Kubernetes worker nodes to interact
with Amazon SQS queues. For simplicity, we will allow all SQS actions here, please do not do so on a
production environment.

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "sqs:*",
            "Resource": "*"
        }
    ]
}
```

## Deploying the Amazon SQS consumer

Now we can start deploying our consumer
```bash
$ kubectl apply -f deploy/consumer-deployment.yaml
```

You can verify the consumer is running by executing this command.
```bash
$ kubectl get deploy
NAME           DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
sqs-consumer   1         1         1            0           5s
```

## Setup Amazon CloudWatch metric and HPA

Next we will need to create an `ExternalMetric` resource for Amazon CloudWatch metric. This resource
will tell the adapter how to retrieve metric data from Amazon CloudWatch. Here in
`deploy/externalmetric.yaml` defined the query parameters used to retrieve the
`ApproximateNumberOfMessagesVisible` for a SQS queue named `helloworld`. For details about how
metric data query works, please refer to [CloudWatch GetMetricData
API](https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html).

```yaml
apiVersion: metrics.aws/v1alpha1
kind: ExternalMetric
metadata:
  name: hello-queue-length
spec:
  name: hello-queue-length
  queries:
    - id: sqs_helloworld
      metricStat:
        metric:
          namespace: "AWS/SQS"
          metricName: "ApproximateNumberOfMessagesVisible"
          dimensions:
            - name: QueueName
              value: "helloworld"
        period: 60
        stat: Average
        unit: Count
      returnData: true
```

Create the ExternalMetric resource
```bash
$ kubectl apply -f deploy/externalmetric.yaml
```

Then setup the HPA for our consumer.
```bash
$ kubectl apply -f deploy/hpa.yaml
```

## Generate load using producer

Finally, we can start generating messages to the queue.

```bash
$ make producer
$ ./bin/producer
```

On a separate terminal, you can now watch your HPA retrieving the queue length and start scaling the
replicas. Amazon SQS now supports metrics at 1-minute interval, so you should start to see the
deployment scale pretty quickly.

```bash
$ kubectl get hpa sqs-consumer-scaler -w
```

## Clean Up

Once you are done with this experiment, you can delete the Kubernetes deployment and respective
resources.

Press `ctrl+c` to terminate the producer if it is still running.

Execute the following commands to remove the consumer, external metric, HPA and Amazon SQS queue.
```bash
$ kubectl delete -f deploy/hpa.yaml
$ kubectl delete -f deploy/externalmetric.yaml
$ kubectl delete -f deploy/consumer-deployment.yaml

$ aws sqs delete-queue helloworld
```

