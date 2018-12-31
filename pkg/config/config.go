package config

type MetricsDiscoveryConfig struct {
	// Series specify how to discover and map CloudWatch metrics to
	// custom metrics API resources.
	Series []MetricSeriesConfig `yaml:"series"`
}

type MetricSeriesConfig struct {
	// Name specifies the series name.
	Name string `yaml:"name"`

	// GroupResource specifies the mapping of this series to a Kubernetes resource.
	Resource GroupResource `yaml:"resource"`

	// Queries specify the CloudWatch metrics query to retrieve data for this series.
	Queries []MetricDataQuery `yaml:"queries"`
}

// GroupResource represents a Kubernetes group-resource.
type GroupResource struct {
	Group    string `yaml:"group,omitempty"`
	Resource string `yaml:"resource"`
}

type MetricDataQuery struct {
	// Resources specifies how associated Kubernetes resources should be discovered for
	// the given metrics.
	Resources string `yaml:"resources"`

	// The math expression to be performed on the returned data, if this structure
	// is performing a math expression. For more information about metric math expressions,
	// see Metric Math Syntax and Functions (http://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/using-metric-math.html#metric-math-syntax)
	// in the Amazon CloudWatch User Guide.
	//
	// Within one MetricDataQuery structure, you must specify either Expression
	// or MetricStat but not both.
	Expression string `yaml:"expression"`

	// A short name used to tie this structure to the results in the response. This
	// name must be unique within a single call to GetMetricData. If you are performing
	// math expressions on this set of data, this name represents that data and
	// can serve as a variable in the mathematical expression. The valid characters
	// are letters, numbers, and underscore. The first character must be a lowercase
	// letter.
	//
	// Id is a required field
	ID string `yaml:"id"`

	// A human-readable label for this metric or expression. This is especially
	// useful if this is an expression, so that you know what the value represents.
	// If the metric or expression is shown in a CloudWatch dashboard widget, the
	// label is shown. If Label is omitted, CloudWatch generates a default.
	Label string `yaml:"label"`

	// The metric to be returned, along with statistics, period, and units. Use
	// this parameter only if this structure is performing a data retrieval and
	// not performing a math expression on the returned data.
	//
	// Within one MetricDataQuery structure, you must specify either Expression
	// or MetricStat but not both.
	MetricStat MetricStat `yaml:"metricStat"`

	// Indicates whether to return the time stamps and raw data values of this metric.
	// If you are performing this call just to do math expressions and do not also
	// need the raw data returned, you can specify False. If you omit this, the
	// default of True is used.
	ReturnData bool `yaml:"returnData"`
}

type MetricStat struct {
	// The metric to return, including the metric name, namespace, and dimensions.
	//
	// Metric is a required field
	Metric Metric `yaml:"metric"`

	// The period to use when retrieving the metric.
	//
	// Period is a required field
	Period int64 `yaml:"period"`

	// The statistic to return. It can include any CloudWatch statistic or extended
	// statistic.
	//
	// Stat is a required field
	Stat string `yaml:"stat"`

	// The unit to use for the returned data points.
	Unit string `yaml:"unit"`
}

type Metric struct {
	// The dimensions for the metric.
	Dimensions []Dimension `yaml:"dimensions"`

	// The name of the metric.
	MetricName string `yaml:"metricName"`

	// The namespace of the metric.
	Namespace string `yaml:"namespace"`
}

type Dimension struct {
	// The name of the dimension.
	//
	// Name is a required field
	Name string `yaml:"name"`

	// The value representing the dimension measurement.
	//
	// Value is a required field
	Value string `yaml:"value"`
}
