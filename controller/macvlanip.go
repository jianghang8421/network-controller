package controller

import (
	"crypto/sha1"
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

// func (c *Controller) enqueueDeletePod(obj interface{}) {
// 	var key string
// 	var err error
// 	if key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj); err != nil {
// 		utilruntime.HandleError(err)
// 		return
// 	}
// 	c.workqueue.Add(key)
// }

func (c *Controller) doAddMacvlanIP(pod *corev1.Pod) error {
	var err error
	if !isMacvlanPod(pod) {
		return nil
	}

	annotationIP := pod.Annotations[macvlanv1.AnnotationIP]
	annotationSubnet := pod.Annotations[macvlanv1.AnnotationSubnet]
	annotationMac := pod.Annotations[macvlanv1.AnnotationMac]

	subnet, err := c.macvlanClientset.MacvlanV1().
		MacvlanSubnets(macvlanv1.MacvlanSubnetNamespace).
		Get(annotationSubnet, metav1.GetOptions{})

	if err != nil {
		c.eventMacvlanSubnetError(pod, err)
		return err
	}

	// allocate ip in subnet
	var allocatedIP net.IP
	var macvlanipCIDR string
	var macvlanipMac string

	if annotationMac == "auto" {
		macvlanipMac = ""
	}

	if annotationIP == "auto" {
		log.Info("alloate ip mode: auto")
		allocatedIP, macvlanipCIDR, err = c.allocateAutoIP(pod, subnet)
	} else if isSingleIP(annotationIP) {
		log.Info("alloate ip mode: single")
		allocatedIP, macvlanipCIDR, err = c.allocateSingleIP(pod, subnet, annotationIP)
	} else if isMultipleIP(annotationIP) {
		log.Info("alloate ip mode: multiple")
		allocatedIP, macvlanipCIDR, macvlanipMac, err = c.allocateMultipleIP(pod, subnet, annotationIP, annotationMac)
	} else {
		c.eventMacvlanIPError(pod, fmt.Errorf("annotation ip invalid: %s", annotationIP))
		return err
	}

	if err != nil {
		c.eventMacvlanIPError(pod, err)
		return err
	}

	// update macvlanip label(ip, selectedip)
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Pod before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := c.kubeClientset.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
		if getErr != nil {
			log.Errorf("Failed to get latest version of Pod: %v", getErr)
			return getErr
		}
		if result.Labels == nil {
			result.Labels = map[string]string{}
		}

		hash := fmt.Sprintf("%x", sha1.Sum([]byte(annotationIP)))
		result.Labels[macvlanv1.LabelMultipleIPHash] = hash
		result.Labels[macvlanv1.LabelSelectedIP] = allocatedIP.String()

		_, updateErr := c.kubeClientset.CoreV1().Pods(result.Namespace).Update(result)

		return updateErr
	})
	if retryErr != nil {
		log.Errorf("pod update labels error: %v", err)
		return err
	}

	// create macvlanip
	macvlanip := makeMacvlanIP(pod, subnet, macvlanipCIDR, macvlanipMac)
	info, err := c.macvlanClientset.MacvlanV1().MacvlanIPs(pod.Namespace).Create(macvlanip)
	if err != nil {
		c.eventMacvlanIPError(pod, err)
		return err
	}

	log.Infof("MacvlanIP created: %v", info.Spec)

	// svc
	if err := c.SyncService(pod); err != nil {
		log.Errorf("Sync service error: %v", err)
	}
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

func makeMacvlanIP(pod *corev1.Pod, subnet *macvlanv1.MacvlanSubnet, cidr, mac string) *macvlanv1.MacvlanIP {
	return &macvlanv1.MacvlanIP{
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
			CIDR:   cidr,
			MAC:    mac,
			PodID:  string(pod.GetUID()),
			Subnet: subnet.Name,
		},
	}
}
