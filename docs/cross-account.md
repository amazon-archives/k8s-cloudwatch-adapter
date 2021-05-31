# Configuring cross account metric example

The `k8s-cloudwatch-adapter` supports retrieving Amazon CloudWatch metrics from another AWS account
using IAM roles. You will assign an IAM role to the `externalmetric` using `roleArn`, the adapter
would then assume this role to send the query to CloudWatch.

Before we start, let's assume the adapter is running with an IAM role called `k8s-cloudwatch-adapter-service-role` 
in Account A. To read the metrics from Account B, you will need to create an IAM role with the
necessary permissions in Account B, let's call this `target-cloudwatch-role`.

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
        },
        {
            "Effect": "Allow",
            "Action": "iam:PassRole",
            "Resource": "arn:aws:iam::<AccountB>:role/target-cloudwatch-role"
        }
    ]
}
```

Make sure we add our `k8s-cloudwatch-adapter-service-role` as a trusted entity so that it is
able to perform `sts:AssumeRole`.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::<AccountA>:role/k8s-cloudwatch-adapter-service-role"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

Next create a file `externalmetric.yaml` and specify the IAM role created in Account B in the 
`roleArn` for your external metric like this

```yaml
apiVersion: metrics.aws/v1alpha1
kind: ExternalMetric
metadata:
  name: sqs-helloworld-length
spec:
  name: sqs-helloworld-length
  roleArn: arn:aws:iam::<AccountB>:role/target-cloudwatch-role
  queries:
    - id: sqs_helloworld_length
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

Deploy the external resource using

```bash
$ kubectl apply -f externalmetric.yaml
```

The adapter should pick this up shortly and start retrieving metrics from CloudWatch in Account B.

For details about the specification of the ExternalMetric custom resource, please check the
 [schema doc](schema.md).