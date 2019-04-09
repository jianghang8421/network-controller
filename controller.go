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

	clientset "github.com/wardenlym/static-pod-controller/pkg/generated/clientset/versioned"
	samplescheme "github.com/wardenlym/static-pod-controller/pkg/generated/clientset/versioned/scheme"
	informers "github.com/wardenlym/static-pod-controller/pkg/generated/informers/externalversions/staticmacvlan/v1"
	listers "github.com/wardenlym/static-pod-controller/pkg/generated/listers/staticmacvlan/v1"
	staticmacvlanv1 "github.com/wardenlym/static-pod-controller/types/apis/staticmacvlan/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const controllerAgentName = "static-pod-controller"

// Controller is the controller implementation for Foo resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	staticmacvlanclientset clientset.Interface

	// deploymentsLister appslisters.DeploymentLister
	// deploymentsSynced cache.InformerSynced

	staticpodsLister listers.StaticPodLister
	staticpodsSynced cache.InformerSynced

	podsLister corelisters.PodLister
	podsSynced cache.InformerSynced
	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	staticmacvlanclientset clientset.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
	podInformer coreinformers.PodInformer,
	staticpodInformer informers.StaticPodInformer) *Controller {

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:          kubeclientset,
		staticmacvlanclientset: staticmacvlanclientset,
		// deploymentsLister:      deploymentInformer.Lister(),
		// deploymentsSynced:      deploymentInformer.Informer().HasSynced,
		staticpodsLister: staticpodInformer.Lister(),
		staticpodsSynced: staticpodInformer.Informer().HasSynced,

		podsLister: podInformer.Lister(),
		podsSynced: podInformer.Informer().HasSynced,

		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "StaticPods"),
		recorder:  recorder,
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Foo resources change
	// staticpodInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	// 	AddFunc: controller.enqueueFoo,
	// 	UpdateFunc: func(old, new interface{}) {
	// 		// controller.enqueueFoo(new)
	// 	},
	// })

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueuePod,
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

func (c *Controller) enqueuePod(obj interface{}) {
	pod := obj.(*corev1.Pod)

	annotationIP, exist := pod.GetAnnotations()["static-ip"]
	if !exist {
		return
	}

	podName := pod.GetName()
	podNamespace := pod.GetNamespace()

	fmt.Printf("Pod add: podName %s - annotationIP %s \n", podName, annotationIP)

	staticpod := &staticmacvlanv1.StaticPod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: podNamespace,
		},
		Spec: staticmacvlanv1.StaticPodSpec{
			ContainerID: "",
			IP:          annotationIP,
			PodID:       string(pod.GetUID()),
			VLan:        "mv123",
		},
	}
	staticpod2, err := c.staticmacvlanclientset.StaticmacvlanV1().StaticPods(podNamespace).Create(staticpod)
	if err != nil {
		fmt.Printf("StaticPods create error: %s %s", err, staticpod2.Name)
	} else {
		fmt.Printf("StaticPods created : %s %s", staticpod2.Name, staticpod2.Spec.IP)
	}
}
