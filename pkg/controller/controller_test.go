package controller

import (
	"errors"
	"testing"

	api "github.com/awslabs/k8s-cloudwatch-adapter/pkg/apis/metrics/v1alpha1"
	"github.com/awslabs/k8s-cloudwatch-adapter/pkg/client/clientset/versioned/fake"
	informers "github.com/awslabs/k8s-cloudwatch-adapter/pkg/client/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type controllerConfig struct {
	// store is the fake etcd backing store that the go client will
	// use to push to the controller.  Add anything the controller to
	// process to the store
	store                      []runtime.Object
	externalMetricsListerCache []*api.ExternalMetric
	syncedFunction             cache.InformerSynced
	enqueuer                   func(c *Controller) func(obj interface{})
	handler                    ControllerHandler
	runtimes                   int
}

type wanted struct {
	keepRunning  bool
	itemsRemaing int

	// number of times added two queue
	// will be zero if the item was forgeten
	enqueCount int
	enquedItem namespacedQueueItem
}

type testConfig struct {
	controllerConfig controllerConfig
	want             wanted
}

func testExternalListerCache() []*api.ExternalMetric {
	var externalMetricsListerCache []*api.ExternalMetric

	externalMetric := newExternalMetric()
	externalMetricsListerCache = append(externalMetricsListerCache, externalMetric)
	return externalMetricsListerCache
}

func TestProcessRunsToCompletionWithExternalMetric(t *testing.T) {
	var storeObjects []runtime.Object
	externalMetric := newExternalMetric()
	storeObjects = append(storeObjects, externalMetric)

	testConfig := testConfig{
		controllerConfig: controllerConfig{
			store:          storeObjects,
			syncedFunction: alwaysSynced,
			handler:        successFakeHandler{},
			runtimes:       1,
		},
		want: wanted{
			itemsRemaing: 0,
			keepRunning:  true,
		},
	}

	runControllerTests(testConfig, t)
}

func TestFailedProcessorReEnqueuesWithExternalMetrics(t *testing.T) {
	var storeObjects []runtime.Object
	externalMetric := newExternalMetric()
	storeObjects = append(storeObjects, externalMetric)

	testConfig := testConfig{
		controllerConfig: controllerConfig{
			store:          storeObjects,
			syncedFunction: alwaysSynced,
			handler:        failedFakeHandler{},
			runtimes:       1,
		},
		want: wanted{
			itemsRemaing: 1,
			keepRunning:  true,
			enqueCount:   2, // should be two because it got added two second time on failure
			enquedItem: namespacedQueueItem{
				namespaceKey: "default/test",
				kind:         "ExternalMetric",
			},
		},
	}

	runControllerTests(testConfig, t)
}

func TestRetryThenRemoveAfter5AttemptsWithExternalMetric(t *testing.T) {
	var storeObjects []runtime.Object
	externalMetric := newExternalMetric()
	storeObjects = append(storeObjects, externalMetric)

	testConfig := testConfig{
		controllerConfig: controllerConfig{
			store:          storeObjects,
			syncedFunction: alwaysSynced,
			handler:        failedFakeHandler{},
			runtimes:       5,
		},
		want: wanted{
			itemsRemaing: 0,
			keepRunning:  true,
			enqueCount:   0, // will be zero after it gets removed
			enquedItem: namespacedQueueItem{
				namespaceKey: "default/test",
				kind:         "ExternalMetric",
			},
		},
	}

	runControllerTests(testConfig, t)
}
func TestInvalidItemOnQueue(t *testing.T) {
	// force the queue to have anything other than a string
	// to exercise the invalid queue path
	var badenquer = func(c *Controller) func(obj interface{}) {
		enquer := func(obj interface{}) {

			// this pushes the object on instead of the key which
			// will cause an error
			c.metricQueue.AddRateLimited(obj)
		}

		return enquer
	}

	var storeObjects []runtime.Object
	externalMetric := newExternalMetric()
	storeObjects = append(storeObjects, externalMetric)

	testConfig := testConfig{
		controllerConfig: controllerConfig{
			store:                      storeObjects,
			syncedFunction:             alwaysSynced,
			enqueuer:                   badenquer,
			handler:                    successFakeHandler{},
			runtimes:                   1,
			externalMetricsListerCache: testExternalListerCache(),
		},
		want: wanted{
			itemsRemaing: 0,
			keepRunning:  true,
		},
	}

	runControllerTests(testConfig, t)
}

func runControllerTests(testConfig testConfig, t *testing.T) {
	c, i := newController(testConfig.controllerConfig)

	stopCh := make(chan struct{})
	defer close(stopCh)
	i.Start(stopCh)

	actaulRunTimes := 0
	keepRunning := false
	for actaulRunTimes < testConfig.controllerConfig.runtimes {
		keepRunning = c.processNextItem()
		actaulRunTimes++
	}

	if actaulRunTimes != testConfig.controllerConfig.runtimes {
		t.Errorf("actual runtime should equal configured runtime = %v, want %v", actaulRunTimes, testConfig.controllerConfig.runtimes)
	}

	if keepRunning != testConfig.want.keepRunning {
		t.Errorf("should continue processing = %v, want %v", keepRunning, testConfig.want.keepRunning)
	}

	items := c.metricQueue.Len()

	if items != testConfig.want.itemsRemaing {
		t.Errorf("Items still on queue = %v, want %v", items, testConfig.want.itemsRemaing)
	}

	retrys := c.metricQueue.NumRequeues(testConfig.want.enquedItem)
	if retrys != testConfig.want.enqueCount {
		t.Errorf("Items enqueued times = %v, want %v", retrys, testConfig.want.enqueCount)
	}
}

func newController(config controllerConfig) (*Controller, informers.SharedInformerFactory) {
	fakeClient := fake.NewSimpleClientset(config.store...)
	i := informers.NewSharedInformerFactory(fakeClient, 0)

	c := NewController(i.Metrics().V1alpha1().ExternalMetrics(), config.handler)

	// override for testing
	c.externalMetricSynced = config.syncedFunction

	if config.enqueuer != nil {
		// override for testings
		c.enqueuer = config.enqueuer(c)
	}

	// override so the item gets added right away for testing with no delay
	c.metricQueue = workqueue.NewNamedRateLimitingQueue(NoDelyRateLimiter(), "nodelay")

	for _, em := range config.externalMetricsListerCache {
		// this will force the enqueuer to reload
		i.Metrics().V1alpha1().ExternalMetrics().Informer().GetIndexer().Add(em)
	}

	return c, i
}

func newExternalMetric() *api.ExternalMetric {
	return &api.ExternalMetric{
		TypeMeta: metav1.TypeMeta{APIVersion: api.SchemeGroupVersion.String(), Kind: "ExternalMetric"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: api.MetricSeriesSpec{
			Name: "test",
			Queries: []api.MetricDataQuery{{
				ID: "queryID",
				MetricStat: api.MetricStat{
					Metric: api.Metric{
						Dimensions: []api.Dimension{{
							Name:  "DimensionName",
							Value: "DimensionValue",
						}},
						MetricName: "metricName",
						Namespace:  "namespace",
					},
					Period: 60,
					Stat:   "Average",
					Unit:   "Count",
				},
				ReturnData: true,
			}},
		},
	}
}

type successFakeHandler struct{}

func (h successFakeHandler) Process(key namespacedQueueItem) error {
	return nil
}

type failedFakeHandler struct{}

func (h failedFakeHandler) Process(key namespacedQueueItem) error {
	return errors.New("this fake always fails")
}

var alwaysSynced = func() bool { return true }

func NoDelyRateLimiter() workqueue.RateLimiter {
	return workqueue.NewItemExponentialFailureRateLimiter(0, 0)
}
