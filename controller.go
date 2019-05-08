package main

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	"github.com/cnrancher/network-controller/cidr"
	clientset "github.com/cnrancher/network-controller/pkg/generated/clientset/versioned"
	macvlanscheme "github.com/cnrancher/network-controller/pkg/generated/clientset/versioned/scheme"
	informers "github.com/cnrancher/network-controller/pkg/generated/informers/externalversions/macvlan/v1"
	listers "github.com/cnrancher/network-controller/pkg/generated/listers/macvlan/v1"
	macvlanv1 "github.com/cnrancher/network-controller/types/apis/macvlan/v1"
	v1 "github.com/cnrancher/network-controller/types/apis/macvlan/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const controllerAgentName = "network-controller"

// Controller is the controller implementation for Foo resources
type Controller struct {
	kubeclientset    kubernetes.Interface
	macvlanclientset clientset.Interface

	macvlanipsLister listers.MacvlanIPLister
	macvlansSynced   cache.InformerSynced

	MacvlanSubnetsLister listers.MacvlanSubnetLister
	MacvlanSubnetsSynced cache.InformerSynced

	podsLister corelisters.PodLister
	podsSynced cache.InformerSynced

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
	recorder  record.EventRecorder
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	macvlanclientset clientset.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
	podInformer coreinformers.PodInformer,
	macvlanipInformer informers.MacvlanIPInformer,
	macvlanSubnetInformer informers.MacvlanSubnetInformer) *Controller {

	utilruntime.Must(macvlanscheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:    kubeclientset,
		macvlanclientset: macvlanclientset,

		macvlanipsLister: macvlanipInformer.Lister(),
		macvlansSynced:   macvlanipInformer.Informer().HasSynced,

		MacvlanSubnetsLister: macvlanSubnetInformer.Lister(),
		MacvlanSubnetsSynced: macvlanSubnetInformer.Informer().HasSynced,

		podsLister: podInformer.Lister(),
		podsSynced: podInformer.Informer().HasSynced,

		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,

		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "network-controller"),
		recorder:  recorder,
	}

	klog.Info("Setting up event handlers")

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.onPodAdd,
		DeleteFunc: controller.onPodDelete,
		UpdateFunc: controller.onPodUpdate,
	})

	macvlanSubnetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.onMacvlanSubnetAdd,
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
	if ok := cache.WaitForCacheSync(stopCh, c.podsSynced, c.macvlansSynced); !ok {
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
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) onPodAdd(obj interface{}) {
	pod := obj.(*corev1.Pod)
	c.doOnPodAdd(pod)
}

func (c *Controller) onPodUpdate(old, new interface{}) {

	oldPod := old.(*corev1.Pod)
	newPod := new.(*corev1.Pod)

	if c.networkSettingChanged(oldPod, newPod) {
		c.doOnPodDelete(oldPod)
		c.doOnPodAdd(newPod)
		// log.Info("do update pod")
		// err := c.kubeclientset.CoreV1().Pods(oldPod.Namespace).Delete(oldPod.Name, &metav1.DeleteOptions{})
		// if err != nil {
		// 	log.Errorf("updating delete error: %v", err)
		// 	return
		// }
		// _, err = c.kubeclientset.CoreV1().Pods(newPod.Namespace).Create(newPod)
		// if err != nil {
		// 	log.Errorf("updating create error: %v", err)
		// 	return
		// }
	}
}

func (c *Controller) networkSettingChanged(old, new *corev1.Pod) bool {

	switch {
	case old.GetAnnotations()["cidr"] != new.GetAnnotations()["cidr"],
		old.GetAnnotations()["subnet"] != new.GetAnnotations()["subnet"],
		old.GetAnnotations()["mac"] != new.GetAnnotations()["mac"]:

		log.Infof(" %s  %s", old.GetAnnotations()["cidr"], new.GetAnnotations()["cidr"])
		log.Infof(" %s  %s", old.GetAnnotations()["subnet"], new.GetAnnotations()["subnet"])
		log.Infof(" %s  %s", old.GetAnnotations()["mac"], new.GetAnnotations()["mac"])
		return true
	default:
		return false
	}
}

func (c *Controller) onPodDelete(obj interface{}) {
	pod := obj.(*corev1.Pod)
	c.doOnPodDelete(pod)
}

func (c *Controller) doOnPodAdd(pod *corev1.Pod) {
	annotationCIDR, exist := pod.GetAnnotations()["cidr"]
	if !exist {
		return
	}

	MacvlanSubnetName, exist := pod.GetAnnotations()["subnet"]
	if !exist {
		return
	}

	podName := pod.GetName()
	podNamespace := pod.GetNamespace()

	vlan, err := c.macvlanclientset.MacvlanV1().MacvlanSubnets("kube-system").Get(MacvlanSubnetName, metav1.GetOptions{})
	if err != nil {
		log.Errorf("Get MacvlanSubnet error: %s %s\n", MacvlanSubnetName, err)
		return
	}

	newAnnotationCIDR, err := func() (string, error) {
		if annotationCIDR == "auto" {
			ips, err := c.macvlanclientset.MacvlanV1().
				MacvlanIPs(podNamespace).
				List(metav1.ListOptions{LabelSelector: "subnet=" + vlan.Name})

			if err != nil {
				return "", fmt.Errorf("list macvlanips error  %s \n", err)

			}
			log.Infof("using ips: %v", ips)
			usedIPs := getUsedIPsFromMacvlanips(ips)
			log.Infof("used ips: %v", ips)
			if len(vlan.Spec.Ranges) == 0 {
				return cidr.AllocateCIDR(vlan.Spec.CIDR, usedIPs)
			}
			// allocate from range
			hosts := CalcHostsFromRanges(vlan.Spec.Ranges)
			return cidr.AllocateInHosts(vlan.Spec.CIDR, hosts, usedIPs)
		}
		return cidr.TryFixNetMask(annotationCIDR, vlan.Spec.CIDR)
	}()
	if err != nil {
		log.Errorf("allocateCIDR error: %s %v", annotationCIDR, err)
	}

	annotationMac := pod.GetAnnotations()["mac"]
	if annotationMac == "auto" {
		annotationMac = ""
	}

	log.Infof("Pod creating: podName %s - annotationIP %s", podName, newAnnotationCIDR)

	macvlanip := &macvlanv1.MacvlanIP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: podNamespace,
			Labels: map[string]string{
				"subnet": vlan.Name,
			},
		},
		Spec: macvlanv1.MacvlanIPSpec{
			CIDR:   newAnnotationCIDR,
			MAC:    annotationMac,
			PodID:  string(pod.GetUID()),
			Subnet: vlan.Name,
		},
	}
	macvlanipInfo, err := c.macvlanclientset.MacvlanV1().MacvlanIPs(podNamespace).Create(macvlanip)
	if err != nil {
		log.Errorf("macvlanips create error: %s %s", err, macvlanipInfo.Name)
	} else {
		log.Infof("macvlanips created : %s %s", macvlanipInfo.Name, macvlanipInfo.Spec.CIDR)
	}
}

func (c *Controller) doOnPodDelete(pod *corev1.Pod) {
	annotationIP, exist := pod.GetAnnotations()["cidr"]
	if !exist {
		return
	}

	podName := pod.GetName()
	podNamespace := pod.GetNamespace()

	log.Infof("Pod delete: podName %s - annotationIP %s ", podName, annotationIP)

	err := c.macvlanclientset.MacvlanV1().MacvlanIPs(podNamespace).Delete(podName, &metav1.DeleteOptions{})
	if err != nil {
		log.Errorf("macvlanips delete error: %s %s", err, podName)
	} else {
		log.Infof("macvlanips deleted : %s", podName)
	}
}

func (c *Controller) enqueuePod(action string, obj *corev1.Pod) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	key = action + "/" + key
	c.workqueue.Add(key)
}

func (c *Controller) onMacvlanSubnetAdd(obj interface{}) {

	subnet, ok := obj.(*macvlanv1.MacvlanSubnet)
	if !ok {
		return
	}

	log.Errorf("MacvlanSubnets Add : %s %v", subnet.Name, subnet)

	if subnet.Labels == nil {
		subnet.Labels = map[string]string{}
	}
	subnet.Labels["master"] = subnet.Spec.Master
	if subnet.Spec.VLAN != 0 {
		subnet.Labels["vlan"] = fmt.Sprint(subnet.Spec.VLAN)
	}

	subnet.Labels["mode"] = subnet.Spec.Mode

	if subnet.Spec.Gateway == "" {
		var err error
		subnet.Spec.Gateway, err = cidr.CalcGatewayByCIDR(subnet.Spec.CIDR)
		if err != nil {
			log.Errorf("CalcGatewayByCIDR error : %v %s", err, subnet.Spec.CIDR)
		}
	}

	_, err := c.macvlanclientset.MacvlanV1().MacvlanSubnets("kube-system").Update(subnet)
	if err != nil {
		log.Errorf("MacvlanSubnets Update : %v %s %v", err, subnet.Name, subnet)
	}
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {

	// Convert the namespace/name string into a distinct namespace and name
	// namespace, name, err := splitActionMetaNamespaceKey(key)
	// if err != nil {
	// 	utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
	// 	return nil
	// }

	return nil
}

func getUsedIPsFromMacvlanips(ips *v1.MacvlanIPList) []string {
	used := []string{}
	for _, item := range ips.Items {
		ip := strings.Split(item.Spec.CIDR, "/")
		if len(ip) == 2 {
			used = append(used, ip[0])
		}
	}
	return used
}

func CalcHostsFromRanges(ranges []v1.IPRange) []string {
	hosts := []string{}

	for _, v := range ranges {
		ips, err := cidr.ParseIPRange(v.RangeStart, v.RangeEnd)
		if err != nil {
			log.Error(err)
		}
		hosts = append(hosts, ips...)
	}

	return RemoveDuplicatesFromSlice(hosts)
}

func RemoveDuplicatesFromSlice(s []string) []string {
	m := make(map[string]bool)
	result := []string{}
	for _, item := range s {
		if _, ok := m[item]; ok {

		} else {
			m[item] = true
			result = append(result, item)
		}
	}
	return result
}
