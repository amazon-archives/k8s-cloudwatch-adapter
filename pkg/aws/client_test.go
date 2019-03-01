package aws

import (
	"fmt"
	"testing"

	"github.com/chankh/k8s-cloudwatch-adapter/pkg/config"
)

func TestQuery(t *testing.T) {
	metricsConfig, err := config.FromFile("testconfig.yaml")
	if err != nil {
		t.Errorf("unable to load metrics discovery configuration: %v", err)
	}

	c := NewCloudWatchClient()
	for _, s := range metricsConfig.Series {
		results, err := c.Query(s.Queries)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		fmt.Printf("Got results: %v\n", results)
	}
}
