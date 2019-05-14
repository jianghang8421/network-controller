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
	podsLister       corelisters.PodLister
	macvlanipLister  listers.MacvlanIPLister
	macvlanclientset clientset.Interface
}

func StartPurgeDaemon(
	macvlanipLister listers.MacvlanIPLister,
	podsLister corelisters.PodLister,
	macvlanclientset clientset.Interface,
	done <-chan struct{}) {

	p := &purger{
		macvlanclientset: macvlanclientset,
		macvlanipLister:  macvlanipLister,
		podsLister:       podsLister,
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
		if p.podNotExist(ip) {
			err = p.macvlanclientset.MacvlanV1().MacvlanIPs(ip.Namespace).Delete(ip.Name, &metav1.DeleteOptions{})
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

func (p *purger) podNotExist(ip *macvlanv1.MacvlanIP) bool {
	_, err := p.podsLister.Pods(ip.Namespace).Get(ip.Name)
	if err != nil && k8serrors.IsNotFound(err) {
		return true
	}
	return false
}
