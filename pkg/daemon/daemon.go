package daemon

import (
	"fmt"
	"os/exec"
	// "strconv"
	"time"

	"github.com/golang/glog"
	ndv1alpha1 "github.com/zshi-redhat/network-device-operator/api/v1alpha1"
	ndclientset "github.com/zshi-redhat/network-device-operator/pkg/client/clientset/versioned"
	ndinformer "github.com/zshi-redhat/network-device-operator/pkg/client/informers/externalversions"
	ndlister "github.com/zshi-redhat/network-device-operator/pkg/client/listers/netdev/v1alpha1"
	"github.com/zshi-redhat/network-device-operator/pkg/utils"
	"golang.org/x/time/rate"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Daemon struct {
	name      string
	stopCh    <-chan struct{}
	k8sClient *kubernetes.Clientset
	ndClient  ndclientset.Interface
	workqueue workqueue.RateLimitingInterface
	ndpLister ndlister.NetDevicePoolLister
}

func New(
	name string,
	stopCh <-chan struct{},
	k8sClient *kubernetes.Clientset,
	ndClient ndclientset.Interface,
) *Daemon {
	return &Daemon{
		name:      name,
		stopCh:    stopCh,
		k8sClient: k8sClient,
		ndClient:  ndClient,
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.NewMaxOfRateLimiter(
			&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(updateDelay), 1)},
			workqueue.NewItemExponentialFailureRateLimiter(1*time.Second, maxUpdateBackoff)), "NetDevicePool"),
	}
}

func (dn *Daemon) Run() error {
	glog.V(2).Info("Run() start")

	// netdevicepool informer
	var timeout int64 = 5
	ndInformerFactory := ndinformer.NewFilteredSharedInformerFactory(dn.ndClient,
		time.Second*15,
		namespace,
		func(lo *metav1.ListOptions) {
			lo.TimeoutSeconds = &timeout
		},
	)
	ndpInformer := ndInformerFactory.Netdev().V1alpha1().NetDevicePools().Informer()
	ndpInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: dn.enqueueNetDevicePoolUpdate,
		UpdateFunc: func(old, new interface{}) {
			dn.enqueueNetDevicePoolUpdate(new)
		},
	})

	dn.ndpLister = ndInformerFactory.Netdev().V1alpha1().NetDevicePools().Lister()

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

func (dn *Daemon) enqueueNetDevicePoolUpdate(obj interface{}) {
	var ok bool
	if _, ok = obj.(*ndv1alpha1.NetDevicePool); !ok {
		utilruntime.HandleError(fmt.Errorf("expected NetDevicePool but got %#v", obj))
		return
	}

	// var key string
	// if (len(ndp.GetNamespace())) > 0 {
	// 	key = ndp.GetNamespace() + "/" + ndp.GetName() + "/" + strconv.FormatInt(ndp.GetGeneration(), 10)
	// } else {
	//	key = ndp.GetName() + "/" + strconv.FormatInt(ndp.GetGeneration(), 10)
	// }
	dn.workqueue.Add(obj)
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
		defer dn.workqueue.Done(obj)

		var ok bool
		var ndp *ndv1alpha1.NetDevicePool
		if ndp, ok = obj.(*ndv1alpha1.NetDevicePool); !ok {
			dn.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected workItem in workqueue but got %#v", obj))
			return nil
		}

		// process workqueue item
		err := dn.syncHandler(ndp)
		if err != nil {
			dn.workqueue.AddRateLimited(obj)
			return fmt.Errorf("Failed to run device sync handler, %v, requeued", err.Error())
		}

		dn.workqueue.Forget(obj)
		glog.V(2).Infof("processWorkItem() successfully processed: %v", ndp)
		return nil
	}(obj)
	if err != nil {
		utilruntime.HandleError(err)
	}

	return true
}

func (dn *Daemon) syncHandler(ndp *ndv1alpha1.NetDevicePool) error {
	glog.V(2).Infof("syncHandler() start")
	_, err := dn.ndpLister.List(labels.Everything())
	if err != nil {
		return err
	}

	_, err = utils.WriteDeviceConfFile(ndp)
	if err != nil {
		return err
	}

	err = configDevice()
	if err != nil {
		return err
	}
	return nil
}

func configDevice() error {
	glog.V(2).Infof("configDevice(): start")
	exit, err := utils.Chroot("/host")
	if err != nil {
		return fmt.Errorf("configDevice(): %v", err)
	}
	defer exit()

	cmd := exec.Command("systemd-run", "--unit", "network-device-daemon-config",
		"--description", fmt.Sprintf("network-device-daemon config device"), "/bin/sh", "-c", "cd /bindata/scripts && drivers.sh")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to config device: %v", err)
	}
	return nil
}

func rebootNode() {
	glog.V(2).Infof("rebootNode(): start")
	exit, err := utils.Chroot("/host")
	if err != nil {
		glog.Errorf("rebootNode(): %v", err)
	}
	defer exit()
	// creates a new transient systemd unit to reboot the system.
	// We explictily try to stop kubelet.service first, before anything else; this
	// way we ensure the rest of system stays running, because kubelet may need
	// to do "graceful" shutdown by e.g. de-registering with a load balancer.
	// However note we use `;` instead of `&&` so we keep rebooting even
	// if kubelet failed to shutdown - that way the machine will still eventually reboot
	// as systemd will time out the stop invocation.
	cmd := exec.Command("systemd-run", "--unit", "network-device-daemon-reboot",
		"--description", fmt.Sprintf("network-device-daemon reboot node"), "/bin/sh", "-c", "systemctl stop kubelet.service; reboot")

	if err := cmd.Run(); err != nil {
		glog.Errorf("failed to reboot node: %v", err)
	}
}
