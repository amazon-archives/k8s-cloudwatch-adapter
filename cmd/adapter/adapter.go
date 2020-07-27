package main

import (
	"flag"
	"os"
	"time"

	"github.com/pkg/errors"
	"k8s.io/component-base/logs"
	"k8s.io/klog"

	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/aws"
	clientset "github.com/awslabs/k8s-cloudwatch-adapter/pkg/client/clientset/versioned"
	informers "github.com/awslabs/k8s-cloudwatch-adapter/pkg/client/informers/externalversions"
	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/controller"
	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/metriccache"
	cwprov "github.com/awslabs/k8s-cloudwatch-adapter/pkg/provider"
	basecmd "github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/cmd"
	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
)

// CloudWatchAdapter represents a custom metrics BaseAdapter for Amazon CloudWatch
type CloudWatchAdapter struct {
	basecmd.AdapterBase
}

func (a *CloudWatchAdapter) makeCloudWatchManager() (aws.CloudWatchManager, error) {
	manager := aws.NewCloudWatchManager()
	return manager, nil
}

func (a *CloudWatchAdapter) newController(cache *metriccache.MetricCache) (*controller.Controller, informers.SharedInformerFactory) {
	clientConfig, err := a.ClientConfig()
	if err != nil {
		klog.Fatalf("unable to construct client config: %v", err)
	}
	adapterClientSet, err := clientset.NewForConfig(clientConfig)
	if err != nil {
		klog.Fatalf("unable to construct lister client to initialize provider: %v", err)
	}

	adapterInformerFactory := informers.NewSharedInformerFactory(adapterClientSet, time.Second*30)
	handler := controller.NewHandler(
		adapterInformerFactory.Metrics().V1alpha1().ExternalMetrics().Lister(),
		cache)

	ctrl := controller.NewController(adapterInformerFactory.Metrics().V1alpha1().ExternalMetrics(), &handler)

	return ctrl, adapterInformerFactory
}

func (a *CloudWatchAdapter) makeProvider(cwManager aws.CloudWatchManager, cache *metriccache.MetricCache) (provider.ExternalMetricsProvider, error) {
	client, err := a.DynamicClient()
	if err != nil {
		return nil, errors.Wrap(err, "unable to construct Kubernetes client")
	}

	mapper, err := a.RESTMapper()
	if err != nil {
		return nil, errors.Wrap(err, "unable to construct RESTMapper")
	}

	cwProvider := cwprov.NewCloudWatchProvider(client, mapper, cwManager, cache)
	return cwProvider, nil
}

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	// set up flags
	cmd := &CloudWatchAdapter{}
	cmd.Name = "k8s-cloudwatch-adapter"
	cmd.Flags().AddGoFlagSet(flag.CommandLine) // make sure we get the klog flags
	cmd.Flags().Parse(os.Args)

	stopCh := make(chan struct{})
	defer close(stopCh)

	cache := metriccache.NewMetricCache()

	// start and run ctrl components
	ctrl, adapterInformerFactory := cmd.newController(cache)
	go adapterInformerFactory.Start(stopCh)
	go ctrl.Run(2, time.Second, stopCh)

	// create CloudWatch client
	cwClient, err := cmd.makeCloudWatchManager()
	if err != nil {
		klog.Fatalf("unable to construct CloudWatch client: %v", err)
	}

	// construct the provider
	cwProvider, err := cmd.makeProvider(cwClient, cache)
	if err != nil {
		klog.Fatalf("unable to construct CloudWatch metrics provider: %v", err)
	}

	cmd.WithExternalMetrics(cwProvider)

	klog.Info("CloudWatch metrics adapter started")

	if err := cmd.Run(stopCh); err != nil {
		klog.Fatalf("unable to run CloudWatch metrics adapter: %v", err)
	}
}
