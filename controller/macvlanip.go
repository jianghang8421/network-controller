package controller

import (
	"fmt"
	"net"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"

	macvlanv1 "github.com/cnrancher/network-controller/types/apis/macvlan/v1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) enqueuePod(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) enqueueDeletePod(obj interface{}) {
	var key string
	var err error
	if key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) addMacvlanIP(pod *corev1.Pod) error {
	var err error
	if !isMacvlanPod(pod) {
		return nil
	}

	annotationIP := pod.Annotations[macvlanv1.AnnotationIP]
	macvlansubnetName := pod.Annotations[macvlanv1.AnnotationSubnet]

	subnet, err := c.macvlanclientset.MacvlanV1().
		MacvlanSubnets(macvlanv1.MacvlanSubnetNamespace).
		Get(macvlansubnetName, metav1.GetOptions{})

	if err != nil {
		c.eventMacvlanSubnetError(pod, err)
		return err
	}

	// allocate ip in subnet
	var allocatedIP net.IP
	var annotationCIDR string
	if annotationIP == "auto" {
		log.Info("alloate in auto")
		allocatedIP, annotationCIDR, err = c.allocateAutoModeIP(pod, subnet)
	} else if isSingleIP(annotationIP) {
		log.Info("alloate in single")
		allocatedIP, annotationCIDR, err = c.allocateSingleIP(pod, subnet, annotationIP)
	} else if isMultipleIP(annotationIP) {
		log.Info("alloate in multiple")
		allocatedIP, annotationCIDR, err = c.allocateMultipleIP(pod, subnet, annotationIP)
	} else {
		c.eventMacvlanIPError(pod, fmt.Errorf("annotation ip invalid: %s", annotationIP))
		return err
	}

	if err != nil {
		c.eventMacvlanIPError(pod, err)
		return err
	}

	// update macvlanip label[ip/selectedip]
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := c.kubeclientset.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
		if getErr != nil {
			log.Errorf("Failed to get latest version of Deployment: %v", getErr)
			return getErr
		}
		if result.Labels == nil {
			result.Labels = map[string]string{}
		}
		result.Labels[macvlanv1.AnnotationSelectedIP] = allocatedIP.String()
		result.Labels[macvlanv1.AnnotationIP] = annotationIP

		_, updateErr := c.kubeclientset.CoreV1().Pods(result.Namespace).Update(result)

		return updateErr
	})
	if retryErr != nil {
		log.Errorf("pod update labels error: %v", err)
		return err
	}

	// create macvlanip
	annotationMac := pod.GetAnnotations()[macvlanv1.AnnotationMac]
	if annotationMac == "auto" {
		annotationMac = ""
	}

	macvlanip := &macvlanv1.MacvlanIP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Labels: map[string]string{
				"subnet": subnet.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					APIVersion: "v1",
					Kind:       "Pod",
					UID:        pod.UID,
					Name:       pod.Name,
				},
			},
		},
		Spec: macvlanv1.MacvlanIPSpec{
			CIDR:   annotationCIDR,
			MAC:    annotationMac,
			PodID:  string(pod.GetUID()),
			Subnet: subnet.Name,
		},
	}
	info, err := c.macvlanclientset.MacvlanV1().MacvlanIPs(pod.Namespace).Create(macvlanip)
	if err != nil {
		c.eventMacvlanIPError(pod, err)
		return err
	}

	log.Infof("MacvlanIP created: %v", info.Spec)
	return nil
}

func isMacvlanPod(pod *corev1.Pod) bool {
	_, exist := pod.GetAnnotations()[macvlanv1.AnnotationIP]
	if !exist {
		return false
	}

	_, exist = pod.GetAnnotations()[macvlanv1.AnnotationSubnet]
	if !exist {
		return false
	}
	return true
}
