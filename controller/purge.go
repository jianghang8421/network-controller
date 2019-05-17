package controller

import (
	"time"

	clientset "github.com/cnrancher/network-controller/pkg/generated/clientset/versioned"
	listers "github.com/cnrancher/network-controller/pkg/generated/listers/macvlan/v1"
	macvlanv1 "github.com/cnrancher/network-controller/types/apis/macvlan/v1"
	log "github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	corelisters "k8s.io/client-go/listers/core/v1"
)

const intervalSeconds int64 = 3600

type purger struct {
	podLister        corelisters.PodLister
	macvlanipLister  listers.MacvlanIPLister
	macvlanClientset clientset.Interface
}

// StartPurgeDaemon purge macvlanips at intervals
func StartPurgeDaemon(
	macvlanipLister listers.MacvlanIPLister,
	podLister corelisters.PodLister,
	macvlanClientset clientset.Interface,
	done <-chan struct{}) {

	p := &purger{
		macvlanClientset: macvlanClientset,
		macvlanipLister:  macvlanipLister,
		podLister:        podLister,
	}
	go wait.JitterUntil(p.purge, time.Duration(intervalSeconds)*time.Second, .1, true, done)
}

func (p *purger) purge() {
	ips, err := p.macvlanipLister.MacvlanIPs("").List(labels.Everything())
	if err != nil {
		log.Errorf("RancherCUBE: error listing tokens during purge: %v", err)
	}

	var count int
	for _, ip := range ips {
		if p.isPodNotExist(ip) {
			err = p.macvlanClientset.MacvlanV1().MacvlanIPs(ip.Namespace).Delete(ip.Name, &metav1.DeleteOptions{})
			if err != nil && !k8serrors.IsNotFound(err) {
				log.Errorf("Purge: error while deleting expired macvlanip %v: %v", err, ip.Name)
				continue
			}
			count++
		}
	}
	if count > 0 {
		log.Infof("Purge: purged %v expired tokens", count)
	}
}

func (p *purger) isPodNotExist(ip *macvlanv1.MacvlanIP) bool {
	_, err := p.podLister.Pods(ip.Namespace).Get(ip.Name)
	if err != nil && k8serrors.IsNotFound(err) {
		return true
	}
	return false
}
