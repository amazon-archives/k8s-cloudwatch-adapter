# External Metric Schema

## ExternalMetric

`ExternalMetric` describes an ExternalMetric resource

Field|Type|Description
---|---|---
kind|string|ExternalMetric
apiVersion|string|metrics.aws/v1alpha1
spec|[MetricSeriesSpec](#metricseriesspec)|Holds all the specifications for this external metric.

## MetricSeriesSpec

`MetricSeriesSpec` contains the specification for a metric series.

Field|Type|Description
---|---|---
name|string|Name of the series
roleArn|string|(Optional) ARN of the IAM role to assume. If specified, the adapter will send requests to Amazon Cloudwatch using this IAM role. 
region|string|(Optional) Target region to retrieve metrics from. The adapter will resolve the current region by default.
queries|[MetricDataQuery](#metricdataquery)[]|Specify the CloudWatch metric queries to retrieve data for this series.

## MetricDataQuery

`MetricDataQuery` represents the query structure used in CloudWatch [GetMetricData](https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html) API.

Field|Type|Description
---|---|---
expression|string|The math expression to be performed on the returned data, if this structure is performing a math expression. For more information about metric math expressions, see [Metric Math Syntax and Functions](http://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/using-metric-math.html#metric-math-syntax) in the Amazon CloudWatch User Guide.<br><br>Within one `MetricDataQuery` structure, you must specify either `Expression` or `MetricStat` but not both.
id|string|A short name used to tie this object to the results in the response. This name must be unique within a single call to GetMetricData. If you are performing math expressions on this set of data, this name represents that data and can serve as a variable in the mathematical expression.<br><br>The valid characters are letters, numbers, and underscore. The first character must be a lowercase letter.
label|string|A human-readable label for this metric or expression. This is especially useful if this is an expression, so that you know what the value represents. If the metric or expression is shown in a CloudWatch dashboard widget, the label is shown. If Label is omitted, CloudWatch generates a default.
metricStat|[MetricStat](#metricstat)|The metric to be returned, along with statistics, period, and units. Use this parameter only if this object is retrieving a metric and not performing a math expression on returned data.<br><br>Within one MetricDataQuery object, you must specify either Expression or MetricStat but not both.
returnData|boolean|Indicates whether to return the timestamps and raw data values of this metric. If you are performing this call just to do math expressions and do not also need the raw data returned, you can specify False. If you omit this, the default of True is used.

## MetricStat

`MetricStat` defines the metric to be returned, along with the statistics, period, and units.

Field|Type|Description
---|---|---
metric|[Metric](#metric)|The metric to return, including the metric name, namespace, and dimensions.
period|integer|The period to use when retrieving the metric.
stat|string|The statistic to return. It can include any CloudWatch statistic or extended statistic.
unit|string|If you omit `unit` then all data that was collected with any unit is returned, along with the corresponding units that were specified when the data was reported to CloudWatch. If you specify a unit, the operation returns only data that was collected with that unit specified. If you specify a unit that does not match the data collected, the results of the operation are null. CloudWatch does not perform unit conversions.

## Metric

`Metric` represents a specific metric.

Field|Type|Description
---|---|---
dimensions|[Dimension](#dimension)[]|The dimensions for the metric.
metricName|string|The name of the metric. This is a required field.
namespace|string|The namespace of the metric.

## Dimension

`Dimension` is a name/value pair that is part of the identity of a metric.

Field|Type|Description
---|---|---
name|string|The name of the dimension. Dimension names cannot contain blank spaces or non-ASCII characters.
value|string|The value of the dimension. Dimension values cannot contain blank spaces or non-ASCII characters.
