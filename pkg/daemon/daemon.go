package daemon

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"golang.org/x/time/rate"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
)

type Daemon struct {
	name      string
	stopCh    <-chan struct{}
	k8sClient *kubernetes.Clientset
	workqueue workqueue.RateLimitingInterface
}

func New(
	name string,
	stopCh <-chan struct{},
	k8sClient *kubernetes.Clientset,
) *Daemon {
	return &Daemon{
		name:      name,
		stopCh:    stopCh,
		k8sClient: k8sClient,
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.NewMaxOfRateLimiter(
			&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(updateDelay), 1)},
			workqueue.NewItemExponentialFailureRateLimiter(1*time.Second, maxUpdateBackoff)), "NetDevicePool"),
	}
}

func (dn *Daemon) Run() error {
	glog.V(2).Info("Run() start")

	// netdevicepool informer

	// worker queue process
	go wait.Until(dn.worker, time.Second, dn.stopCh)

	for {
		select {
		case <-dn.stopCh:
			glog.V(2).Info("Run() stop")
			return nil
		}
	}
}

func (dn *Daemon) worker() {
	for dn.processWorkItem() {
	}
}

func (dn *Daemon) processWorkItem() bool {
	glog.V(2).Infof("processWorkItem() start")
	glog.V(2).Infof("processWorkItem() work queue size: %d", dn.workqueue.Len())

	obj, shutdown := dn.workqueue.Get()
	if shutdown {
		glog.V(2).Infof("processWorkItem() shutdown workqueue")
		return false
	}

	err := func(obj interface{}) error {
		var key int64
		var ok bool
		defer dn.workqueue.Done(obj)
		if key, ok = obj.(int64); !ok {
			dn.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected workItem in workqueue but got %#v", obj))
			return nil
		}

		// process workqueue item
		err := dn.syncHandler(key)
		if err != nil {
			dn.workqueue.AddRateLimited(key)
			return fmt.Errorf("Failed to run device sync handler, %v, requeued", err.Error())
		}

		dn.workqueue.Forget(obj)
		glog.V(2).Infof("processWorkItem() successfully processed: %d", key)
		return nil
	}(obj)
	if err != nil {
		utilruntime.HandleError(err)
	}

	return true
}

func (dn *Daemon) syncHandler(key int64) error {
	glog.V(2).Infof("syncHandler() start")
	return nil
}
