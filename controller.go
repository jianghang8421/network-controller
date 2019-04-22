package main

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	"github.com/cnrancher/static-pod-controller/cidr"
	clientset "github.com/cnrancher/static-pod-controller/pkg/generated/clientset/versioned"
	staticmacvlanscheme "github.com/cnrancher/static-pod-controller/pkg/generated/clientset/versioned/scheme"
	informers "github.com/cnrancher/static-pod-controller/pkg/generated/informers/externalversions/staticmacvlan/v1"
	listers "github.com/cnrancher/static-pod-controller/pkg/generated/listers/staticmacvlan/v1"
	staticmacvlanv1 "github.com/cnrancher/static-pod-controller/types/apis/staticmacvlan/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const controllerAgentName = "static-pod-controller"

// Controller is the controller implementation for Foo resources
type Controller struct {
	kubeclientset          kubernetes.Interface
	staticmacvlanclientset clientset.Interface

	staticpodsLister listers.StaticPodLister
	staticpodsSynced cache.InformerSynced

	vlansubnetsLister listers.VLANSubnetLister
	vlansubnetsSynced cache.InformerSynced

	podsLister corelisters.PodLister
	podsSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
	recorder  record.EventRecorder
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	staticmacvlanclientset clientset.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
	podInformer coreinformers.PodInformer,
	staticpodInformer informers.StaticPodInformer,
	vlansubnetInformer informers.VLANSubnetInformer) *Controller {

	utilruntime.Must(staticmacvlanscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:          kubeclientset,
		staticmacvlanclientset: staticmacvlanclientset,

		staticpodsLister: staticpodInformer.Lister(),
		staticpodsSynced: staticpodInformer.Informer().HasSynced,

		vlansubnetsLister: vlansubnetInformer.Lister(),
		vlansubnetsSynced: vlansubnetInformer.Informer().HasSynced,

		podsLister: podInformer.Lister(),
		podsSynced: podInformer.Informer().HasSynced,

		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "StaticPods"),
		recorder:  recorder,
	}

	klog.Info("Setting up event handlers")

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.onPodAdd,
		DeleteFunc: controller.onPodDelete,
		UpdateFunc: controller.onPodUpdate,
	})

	vlansubnetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.onVLANSubnetAdd,
		DeleteFunc: controller.onVLANSubnetDelete,
		UpdateFunc: controller.onVLANSubnetUpdate,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Staticmacvlan controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.podsSynced, c.staticpodsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	_, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	return true
}

func (c *Controller) onPodAdd(obj interface{}) {
	pod := obj.(*corev1.Pod)

	annotationIP, exist := pod.GetAnnotations()["static-ip"]
	if !exist {
		return
	}

	vlansubnetName, exist := pod.GetAnnotations()["vlan"]
	if !exist {
		return
	}

	podName := pod.GetName()
	podNamespace := pod.GetNamespace()

	// TODO: use kube-system or cattle-system namesapce
	vlan, err := c.staticmacvlanclientset.StaticmacvlanV1().VLANSubnets("kube-system").Get(vlansubnetName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Get VLANSubnet error: %s %s\n", vlansubnetName, err)
		return
	}

	if annotationIP == "auto" {
		annotationIP, err = cidr.RandomCIDR(vlan.Spec.CIDR)
		if err != nil {
			fmt.Printf("allocteIPinVLAN  %s  cidr(%s)  ip(%s)\n", err, vlan.Spec.CIDR, annotationIP)
			return
		}
	}

	annotationMac := pod.GetAnnotations()["static-mac"]

	fmt.Printf("Pod creating: podName %s - annotationIP %s \n", podName, annotationIP)

	staticpod := &staticmacvlanv1.StaticPod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: podNamespace,
		},
		Spec: staticmacvlanv1.StaticPodSpec{
			IP:    annotationIP,
			MAC:   annotationMac,
			PodID: string(pod.GetUID()),
			VLAN:  vlan.Name,
		},
	}
	staticpodInfo, err := c.staticmacvlanclientset.StaticmacvlanV1().StaticPods(podNamespace).Create(staticpod)
	if err != nil {
		fmt.Printf("StaticPods create error: %s %s", err, staticpodInfo.Name)
	} else {
		fmt.Printf("StaticPods created : %s %s", staticpodInfo.Name, staticpodInfo.Spec.IP)
	}
}

func (c *Controller) onPodUpdate(old, new interface{}) {
}

func (c *Controller) onPodDelete(obj interface{}) {

	pod := obj.(*corev1.Pod)

	annotationIP, exist := pod.GetAnnotations()["static-ip"]
	if !exist {
		return
	}

	podName := pod.GetName()
	podNamespace := pod.GetNamespace()

	fmt.Printf("Pod delete: podName %s - annotationIP %s \n", podName, annotationIP)

	err := c.staticmacvlanclientset.StaticmacvlanV1().StaticPods(podNamespace).Delete(podName, &metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("StaticPods delete error: %s %s", err, podName)
	} else {
		fmt.Printf("StaticPods deleted : %s", podName)
	}
}

func (c *Controller) onVLANSubnetAdd(obj interface{}) {

}

func (c *Controller) onVLANSubnetUpdate(old, new interface{}) {

}

func (c *Controller) onVLANSubnetDelete(obj interface{}) {

}
