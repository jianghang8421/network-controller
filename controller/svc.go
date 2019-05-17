package controller

import (
	"fmt"
	"strings"

	macvlanv1 "github.com/cnrancher/network-controller/types/apis/macvlan/v1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SyncService auto create/delete svc for macvlan pod
func (c *Controller) SyncService(pod *corev1.Pod, macvlanip *macvlanv1.MacvlanIP) error {
	// do nothing if pod has no owner
	if pod.OwnerReferences == nil || len(pod.OwnerReferences) == 0 {
		return nil
	}

	ownerName, ownerKind, err := c.findOwnerWorkload(pod)
	if err != nil {
		return err
	}
	log.Infof("%s is own by workload %s", pod.Name, ownerName)

	svc, err := c.kubeClientset.CoreV1().Services(pod.Namespace).Get(ownerName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	macvlanService := makeService(ownerKind, svc.Namespace, fmt.Sprintf("%s-macvlan", svc.Name))
	err = c.createServiceIfNotExist(macvlanService)
	if err != nil {
		return err
	}
	err = c.updateEndpoint(macvlanService, pod, macvlanip)
	return err
}

func (c *Controller) findOwnerWorkload(pod *corev1.Pod) (string, string, error) {
	return func() (string, string, error) {
		for _, owner := range pod.OwnerReferences {
			switch owner.Kind {
			case "DaemonSet", "StatefulSet", "Deployment":
				return c.getAppName(owner.Kind, pod.Namespace, owner.Name)
			case "ReplicaSet":
				rs, err := c.kubeClientset.AppsV1().
					ReplicaSets(pod.Namespace).
					Get(owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", "", err
				}
				if rs.OwnerReferences == nil ||
					len(rs.OwnerReferences) < 1 ||
					rs.OwnerReferences[0].Kind != "Deployment" {
					return "", "", fmt.Errorf("pod owner is invalid kind: %s", rs.OwnerReferences[0].Kind)
				}
				return c.getAppName(rs.OwnerReferences[0].Kind, "Deployment", owner.Name)
			}
		}
		return "", "", fmt.Errorf("%s owner workload not found", pod.Name)
	}()
}

func (c *Controller) getAppName(kind, namespace, name string) (string, string, error) {
	rcs := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: strings.ToLower(kind) + "s",
	}
	app, err := c.kubeDynamicClientset.
		Resource(rcs).
		Namespace(namespace).
		Get(name, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	return app.GetName(), app.GetKind(), nil
}

func makeService(kind, namespace, name string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					APIVersion: "v1",
					Kind:       kind,
					Name:       name,
				},
			},
		},
		Spec: corev1.ServiceSpec{},
	}
}

func (c *Controller) createServiceIfNotExist(svc *corev1.Service) error {
	_, err := c.kubeClientset.CoreV1().Services(svc.Namespace).Get(svc.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = c.kubeClientset.CoreV1().Services(svc.Namespace).Create(svc)
			return err
		}
	}
	return err
}

func (c *Controller) updateEndpoint(svc *corev1.Service, pod *corev1.Pod, macvlanip *macvlanv1.MacvlanIP) error {

	// get endpoint or create if not exist
	get := func() (*corev1.Endpoints, error) {
		ep, err := c.kubeClientset.CoreV1().Endpoints(svc.Namespace).Get(svc.Name, metav1.GetOptions{})
		if err == nil {
			return ep, nil
		}
		if errors.IsNotFound(err) {
			ep, err = c.kubeClientset.CoreV1().Endpoints(svc.Namespace).
				Create(&corev1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						Name:      svc.Name,
						Namespace: svc.Namespace,
					}})
			if err != nil {
				return nil, err
			}
			return ep, nil
		}
		return nil, err
	}

	ep, err := get()
	if err != nil {
		return err
	}

	ip := ""
	a := strings.Split(macvlanip.Spec.CIDR, "/")
	if len(a) == 2 {
		ip = a[0]
	}

	for _, v := range ep.Subsets {
		for _, vv := range v.Addresses {
			if vv.IP == ip {
				return nil
			}
		}
	}

	// ???

	return nil
}
