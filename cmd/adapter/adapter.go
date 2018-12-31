package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/util/logs"

	"github.com/chankh/k8s-cloudwatch-adapter/pkg/aws"
	adaptercfg "github.com/chankh/k8s-cloudwatch-adapter/pkg/config"
	cwprov "github.com/chankh/k8s-cloudwatch-adapter/pkg/provider"
	basecmd "github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/cmd"
	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
)

type CloudWatchAdapter struct {
	basecmd.AdapterBase

	// AWSRegion is the region where CloudWatch metrics are pulled.
	AWSRegion string

	// AdapterConfigFile points to the file containing the metrics discovery configuration.
	AdapterConfigFile string

	metricsConfig *adaptercfg.MetricsDiscoveryConfig
}

func (a *CloudWatchAdapter) makeCloudWatchClient() (aws.Client, error) {
	client := aws.NewCloudWatchClient()
	return client, nil
}

func (a *CloudWatchAdapter) addFlags() {
	a.Flags().StringVar(&a.AdapterConfigFile, "config", a.AdapterConfigFile,
		"Configuration file containing CloudWatch metric queries")
}

func (a *CloudWatchAdapter) loadConfig() error {
	// load metrics discovery configuration
	if a.AdapterConfigFile == "" {
		return fmt.Errorf("no metrics discovery configuration file specified (make sure to use --config)")
	}
	metricsConfig, err := adaptercfg.FromFile(a.AdapterConfigFile)
	if err != nil {
		return fmt.Errorf("unable to load metrics discovery configuration, %v", err)
	}

	a.metricsConfig = metricsConfig

	return nil
}

func (a *CloudWatchAdapter) makeProvider(cwClient aws.Client) (provider.MetricsProvider, error) {
	if len(a.metricsConfig.Series) == 0 {
		return nil, errors.New("no metric series configured")
	}

	client, err := a.DynamicClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to construct Kubernetes client")
	}

	mapper, err := a.RESTMapper()
	if err != nil {
		return nil, errors.Wrap(err, "unable to construct RESTMapper")
	}

	cwProvider := cwprov.NewCloudWatchProvider(client, mapper, cwClient, a.metricsConfig.Series)
	return cwProvider, nil
}

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	// set up flags
	cmd := &CloudWatchAdapter{}
	cmd.Name = "cloudwatch-metrics-adapter"
	cmd.addFlags()
	cmd.Flags().AddGoFlagSet(flag.CommandLine) // make sure we get the glog flags
	cmd.Flags().Parse(os.Args)

	// create CloudWatch client
	cwClient, err := cmd.makeCloudWatchClient()
	if err != nil {
		glog.Fatalf("unable to construct CloudWatch client: %v", err)
	}

	// load the config
	if err := cmd.loadConfig(); err != nil {
		glog.Fatalf("unable to load config: %v", err)
	}

	// construct the provider
	cwProvider, err := cmd.makeProvider(cwClient)
	if err != nil {
		glog.Fatalf("unable to construct CloudWatch metrics provider: %v", err)
	}

	cmd.WithCustomMetrics(cwProvider)
	cmd.WithExternalMetrics(cwProvider)

	glog.Info("CloudWatch metrics adapter started")

	if err := cmd.Run(wait.NeverStop); err != nil {
		glog.Fatalf("unable to run custom metrics adapter: %v", err)
	}
}
