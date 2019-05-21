package controller

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/intstr"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// SyncService auto create/delete svc for macvlan pod
func (c *Controller) SyncService(name, namespace string) error {

	pod, err := c.kubeClientset.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// do nothing if pod has no owner
	if pod.OwnerReferences == nil || len(pod.OwnerReferences) == 0 {
		return nil
	}

	ownerName, ownerKind, ownerUID, err := c.findOwnerWorkload(pod)
	if err != nil {
		return err
	}
	log.Infof("%s is own by workload %s", pod.Name, ownerName)

	svc, err := c.kubeClientset.CoreV1().Services(pod.Namespace).Get(ownerName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	macvlanService := makeService(ownerUID, ownerKind, svc)
	err = c.createServiceIfNotExist(macvlanService)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) findOwnerWorkload(pod *corev1.Pod) (string, string, types.UID, error) {
	return func() (string, string, types.UID, error) {
		for _, owner := range pod.OwnerReferences {
			switch owner.Kind {
			case "DaemonSet", "StatefulSet", "Deployment":
				return c.getAppName(owner.Kind, pod.Namespace, owner.Name)
			case "ReplicaSet":
				rs, err := c.kubeClientset.AppsV1().
					ReplicaSets(pod.Namespace).
					Get(owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", "", "", err
				}
				if rs.OwnerReferences == nil ||
					len(rs.OwnerReferences) < 1 ||
					rs.OwnerReferences[0].Kind != "Deployment" {
					return "", "", "", fmt.Errorf("pod owner is invalid kind: %s", rs.OwnerReferences[0].Kind)
				}
				return c.getAppName("Deployment", pod.Namespace, rs.OwnerReferences[0].Name)
			}
		}
		return "", "", "", fmt.Errorf("%s owner workload not found", pod.Name)
	}()
}

func (c *Controller) getAppName(kind, namespace, name string) (string, string, types.UID, error) {
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
		return "", "", "", err
	}
	return app.GetName(), app.GetKind(), app.GetUID(), nil
}

func makeService(uid types.UID, kind string, svc *corev1.Service) *corev1.Service {

	ports := []corev1.ServicePort{}

	for _, v := range svc.Spec.Ports {
		port := v.DeepCopy()
		port.Port = port.Port + 1
		port.TargetPort = intstr.FromInt(port.TargetPort.IntValue() + 1)
		ports = append(ports, *port)
	}

	spec := svc.Spec.DeepCopy()
	spec.Ports = ports

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-macvlan", svc.Name),
			Namespace:       svc.Namespace,
			OwnerReferences: svc.OwnerReferences,
			Annotations: map[string]string{
				"k8s.v1.cni.cncf.io/networks": netAttatchDefName,
			},
		},
		Spec: *spec,
	}

	return s
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
