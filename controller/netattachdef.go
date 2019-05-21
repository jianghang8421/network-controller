package controller

import (
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// TODO: rename to "static-macvlan-cni-attach"
	netAttatchDefName = "static-macvlan-cni"
	netAttatchDef     = schema.GroupVersionResource{
		Group:    "k8s.cni.cncf.io",
		Version:  "v1",
		Resource: "network-attachment-definitions",
	}
)

func makeNetworkAttachmentDefinition(name, namespace string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "NetworkAttachmentDefinition",
			"apiVersion": netAttatchDef.Group + "/" + netAttatchDef.Version,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"config": `{
					"cniVersion": "0.3.0",
					"type": "static-macvlan-cni",
					"master": "",
					"ipam": {
					  "type": "static-ipam",
					  "addresses": [
						  {
							"address": "{address}",
							"gateway": "{gateway}"
						  }
						]
					}
				  }`,
			},
		},
	}
}

func (c *Controller) onNamespaceAdd(obj interface{}) {
	ns, ok := obj.(*corev1.Namespace)
	if !ok {
		return
	}
	log.Infof("Namespace created: %s %s", ns.Namespace, ns.Name)

	_, err := c.kubeDynamicClientset.
		Resource(netAttatchDef).
		Namespace(ns.Name).
		Create(makeNetworkAttachmentDefinition(netAttatchDefName, ns.Name), metav1.CreateOptions{})
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			log.Infof("NetworkAttachmentDef create error: %s %v", ns.Name, err)
		}
	}
}

func (c *Controller) onNamespaceDelete(obj interface{}) {
	ns, ok := obj.(*corev1.Namespace)
	if !ok {
		return
	}
	log.Infof("Namespace delete: %s %s", ns.Namespace, ns.Name)

	err := c.kubeDynamicClientset.
		Resource(netAttatchDef).
		Namespace(ns.Name).
		Delete(netAttatchDefName, &metav1.DeleteOptions{})
	if err != nil {
		log.Infof("NetworkAttachmentDef delete error: %s %v", ns.Name, err)
	}
}
