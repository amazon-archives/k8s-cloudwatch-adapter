package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apiserver/pkg/util/logs"

	"github.com/chankh/k8s-cloudwatch-adapter/pkg/aws"
	clientset "github.com/chankh/k8s-cloudwatch-adapter/pkg/client/clientset/versioned"
	informers "github.com/chankh/k8s-cloudwatch-adapter/pkg/client/informers/externalversions"
	"github.com/chankh/k8s-cloudwatch-adapter/pkg/config"
	adaptercfg "github.com/chankh/k8s-cloudwatch-adapter/pkg/config"
	"github.com/chankh/k8s-cloudwatch-adapter/pkg/controller"
	"github.com/chankh/k8s-cloudwatch-adapter/pkg/metriccache"
	cwprov "github.com/chankh/k8s-cloudwatch-adapter/pkg/provider"
	basecmd "github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/cmd"
	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
)

type CloudWatchAdapter struct {
	basecmd.AdapterBase

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

func (a *CloudWatchAdapter) newController(metriccache *metriccache.MetricCache) (*controller.Controller, informers.SharedInformerFactory) {
	clientConfig, err := a.ClientConfig()
	if err != nil {
		glog.Fatalf("unable to construct client config: %v", err)
	}
	adapterClientSet, err := clientset.NewForConfig(clientConfig)
	if err != nil {
		glog.Fatalf("unable to construct lister client to initialize provider: %v", err)
	}

	adapterInformerFactory := informers.NewSharedInformerFactory(adapterClientSet, time.Second*30)
	handler := controller.NewHandler(
		adapterInformerFactory.Metrics().V1alpha1().ExternalMetrics().Lister(),
		metriccache)

	controller := controller.NewController(adapterInformerFactory.Metrics().V1alpha1().ExternalMetrics(), &handler)

	return controller, adapterInformerFactory
}
func (a *CloudWatchAdapter) loadConfig() error {
	// config file is optional
	if a.AdapterConfigFile == "" {
		a.metricsConfig = &config.MetricsDiscoveryConfig{}
		return nil
	}

	// load metrics discovery configuration
	metricsConfig, err := adaptercfg.FromFile(a.AdapterConfigFile)
	if err != nil {
		return fmt.Errorf("unable to load metrics discovery configuration, %v", err)
	}

	a.metricsConfig = metricsConfig

	return nil
}

func (a *CloudWatchAdapter) makeProvider(cwClient aws.Client, metriccache *metriccache.MetricCache) (provider.MetricsProvider, error) {
	client, err := a.DynamicClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to construct Kubernetes client")
	}

	mapper, err := a.RESTMapper()
	if err != nil {
		return nil, errors.Wrap(err, "unable to construct RESTMapper")
	}

	cwProvider := cwprov.NewCloudWatchProvider(client, mapper, cwClient, a.metricsConfig.Series, metriccache)
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

	stopCh := make(chan struct{})
	defer close(stopCh)

	metriccache := metriccache.NewMetricCache()

	// start and run contoller components
	controller, adapterInformerFactory := cmd.newController(metriccache)
	go adapterInformerFactory.Start(stopCh)
	go controller.Run(2, time.Second, stopCh)

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
	cwProvider, err := cmd.makeProvider(cwClient, metriccache)
	if err != nil {
		glog.Fatalf("unable to construct CloudWatch metrics provider: %v", err)
	}

	cmd.WithCustomMetrics(cwProvider)
	cmd.WithExternalMetrics(cwProvider)

	glog.Info("CloudWatch metrics adapter started")

	if err := cmd.Run(stopCh); err != nil {
		glog.Fatalf("unable to run CloudWatch metrics adapter: %v", err)
	}
}
