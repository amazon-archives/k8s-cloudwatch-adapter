package aws

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go/aws"
	api "github.com/awslabs/k8s-cloudwatch-adapter/pkg/apis/metrics/v1alpha1"
)

func TestToCloudWatchQuery(t *testing.T) {
	externalMetric := newFullExternalMetric("test")
	metricRequest := toCloudWatchQuery(externalMetric)

	// Metric Queries
	if len(metricRequest.MetricDataQueries) != len(externalMetric.Spec.Queries) {
		t.Errorf("metricRequest Queries = %v, want %v", metricRequest.MetricDataQueries, externalMetric.Spec.Queries)
	}

	for i, q := range metricRequest.MetricDataQueries {
		wantQueries := externalMetric.Spec.Queries[i]
		if q.Expression != nil && *q.Expression != wantQueries.Expression {
			t.Errorf("metricRequest Expression = %v, want %v", q.Expression, wantQueries.Expression)
		}

		if *q.Id != wantQueries.ID {
			t.Errorf("metricRequest ID = %v, want %v", q.Id, wantQueries.ID)
		}

		if *q.Label != wantQueries.Label {
			t.Errorf("metricRequest Label = %v, want %v", q.Label, wantQueries.Label)
		}

		qStat := q.MetricStat
		wantStat := wantQueries.MetricStat

		if qStat != nil {
			qMetric := qStat.Metric
			wantMetric := wantStat.Metric

			if len(qMetric.Dimensions) != len(wantMetric.Dimensions) {
				t.Errorf("metricRequest Dimensions = %v, want = %v", qMetric.Dimensions, wantMetric.Dimensions)
			}

			for j, d := range qMetric.Dimensions {
				if *d.Name != wantMetric.Dimensions[j].Name {
					t.Errorf("metricRequest Dimension Name = %v, want = %v", *d.Name, wantMetric.Dimensions[j].Name)
				}

				if *d.Value != wantMetric.Dimensions[j].Value {
					t.Errorf("metricRequest Dimension Value = %v, want = %v", *d.Value, wantMetric.Dimensions[j].Value)
				}
			}

			if *qMetric.MetricName != wantMetric.MetricName {
				t.Errorf("metricRequest MetricName = %v, want %v", *qMetric.MetricName, wantMetric.MetricName)
			}

			if *qMetric.Namespace != wantMetric.Namespace {
				t.Errorf("metricRequest Namespace = %v, want %v", *qMetric.Namespace, wantMetric.Namespace)
			}

			if *qStat.Period != wantStat.Period {
				t.Errorf("metricRequest Period = %v, want %v", *qStat.Period, wantStat.Period)
			}

			if *qStat.Stat != wantStat.Stat {
				t.Errorf("metricRequest Stat = %v, want %v", *qStat.Stat, wantStat.Stat)
			}

			if aws.StringValue(qStat.Unit) != wantStat.Unit {
				t.Errorf("metricRequest Unit = %v, want %v", qStat.Unit, wantStat.Unit)
			}
		}

		if *q.ReturnData != wantQueries.ReturnData {
			t.Errorf("metricRequest ReturnData = %v, want %v", *q.ReturnData, wantQueries.ReturnData)
		}
	}
}

func newFullExternalMetric(name string) *api.ExternalMetric {
	return &api.ExternalMetric{
		TypeMeta: metav1.TypeMeta{APIVersion: api.SchemeGroupVersion.String(), Kind: "ExternalMetric"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: api.MetricSeriesSpec{
			Name:    "Name",
			RoleARN: "MyRoleARN",
			Region:  "Region",
			Queries: []api.MetricDataQuery{
				{
					ID:         "query1",
					Expression: "query2/query3",
				},
				{
					ID: "query2",
					MetricStat: api.MetricStat{
						Metric: api.Metric{
							Dimensions: []api.Dimension{{
								Name:  "DimensionName1",
								Value: "DimensionValue1",
							}},
							MetricName: "metricName1",
							Namespace:  "namespace1",
						},
						Period: 60,
						Stat:   "Average",
						Unit:   "Bytes",
					},
					ReturnData: true,
				},
				{
					ID: "query3",
					MetricStat: api.MetricStat{
						Metric: api.Metric{
							Dimensions: []api.Dimension{{
								Name:  "DimensionName2",
								Value: "DimensionValue2",
							},
								{
									Name:  "DimensionName3",
									Value: "DimensionValue3",
								}},
							MetricName: "metricName2",
							Namespace:  "namespace2",
						},
						Period: 60,
						Stat:   "Sum",
						Unit:   "Count",
					},
					ReturnData: false,
				},
			},
		},
	}
}
