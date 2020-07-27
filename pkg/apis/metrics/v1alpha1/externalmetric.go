package v1alpha1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +genclient:skipVerbs=patch
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExternalMetric describes a ExternalMetric resource
type ExternalMetric struct {
	// TypeMeta is the metadata for the resource, like kind and apiversion
	meta_v1.TypeMeta `json:",inline"`

	// ObjectMeta contains the metadata for the particular object (name, namespace, self link,
	// labels, etc)
	meta_v1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the custom resource spec
	Spec MetricSeriesSpec `json:"spec"`
}

// MetricSeriesSpec contains the specification for a metric series.
type MetricSeriesSpec struct {
	// Name specifies the series name.
	Name string `json:"name"`

	// RoleARN indicate the ARN of IAM role to assume, this metric will be retrieved using this role.
	RoleARN string `json:"roleArn"`

	// Region specifies the region where metrics should be retrieved.
	Region string `json:"region"`

	// Queries specify the CloudWatch metrics query to retrieve data for this series.
	Queries []MetricDataQuery `json:"queries"`
}

// MetricDataQuery represents the query structure used in GetMetricData operation to CloudWatch API.
type MetricDataQuery struct {
	// Resources specifies how associated Kubernetes resources should be discovered for
	// the given metrics.
	Resources string `json:"resources"`

	// The math expression to be performed on the returned data, if this structure
	// is performing a math expression. For more information about metric math expressions,
	// see Metric Math Syntax and Functions (http://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/using-metric-math.html#metric-math-syntax)
	// in the Amazon CloudWatch User Guide.
	//
	// Within one MetricDataQuery structure, you must specify either Expression
	// or MetricStat but not both.
	Expression string `json:"expression"`

	// A short name used to tie this structure to the results in the response. This
	// name must be unique within a single call to GetMetricData. If you are performing
	// math expressions on this set of data, this name represents that data and
	// can serve as a variable in the mathematical expression. The valid characters
	// are letters, numbers, and underscore. The first character must be a lowercase
	// letter.
	//
	// Id is a required field
	ID string `json:"id"`

	// A human-readable label for this metric or expression. This is especially
	// useful if this is an expression, so that you know what the value represents.
	// If the metric or expression is shown in a CloudWatch dashboard widget, the
	// label is shown. If Label is omitted, CloudWatch generates a default.
	Label string `json:"label"`

	// The metric to be returned, along with statistics, period, and units. Use
	// this parameter only if this structure is performing a data retrieval and
	// not performing a math expression on the returned data.
	//
	// Within one MetricDataQuery structure, you must specify either Expression
	// or MetricStat but not both.
	MetricStat MetricStat `json:"metricStat"`

	// Indicates whether to return the time stamps and raw data values of this metric.
	// If you are performing this call just to do math expressions and do not also
	// need the raw data returned, you can specify False. If you omit this, the
	// default of True is used.
	ReturnData bool `json:"returnData"`
}

// MetricStat defines the metric to be returned, along with the statistics, period, and units.
type MetricStat struct {
	// The metric to return, including the metric name, namespace, and dimensions.
	//
	// Metric is a required field
	Metric Metric `json:"metric"`

	// The period to use when retrieving the metric.
	//
	// Period is a required field
	Period int64 `json:"period"`

	// The statistic to return. It can include any CloudWatch statistic or extended
	// statistic.
	//
	// Stat is a required field
	Stat string `json:"stat"`

	// The unit to use for the returned data points.
	Unit string `json:"unit"`
}

// Metric represents a specific metric.
type Metric struct {
	// The dimensions for the metric.
	Dimensions []Dimension `json:"dimensions"`

	// The name of the metric.
	MetricName string `json:"metricName"`

	// The namespace of the metric.
	Namespace string `json:"namespace"`
}

// Dimension expands the identity of a metric.
type Dimension struct {
	// The name of the dimension.
	//
	// Name is a required field
	Name string `json:"name"`

	// The value representing the dimension measurement.
	//
	// Value is a required field
	Value string `json:"value"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExternalMetricList is a list of ExternalMetric resources
type ExternalMetricList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`

	Items []ExternalMetric `json:"items"`
}
