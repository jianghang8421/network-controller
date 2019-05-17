package controller

import (
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	networkAttatchmentConfigName = "static-macvlan-cni-cfg"
	networkAttatchmentDefinition = schema.GroupVersionResource{
		Group:    "k8s.cni.cncf.io",
		Version:  "v1",
		Resource: "network-attachment-definitions",
	}
)

func makeNetworkAttachmentDefinition(name, namespace string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "NetworkAttachmentDefinition",
			"apiVersion": networkAttatchmentDefinition.Group + "/" + networkAttatchmentDefinition.Version,
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
		Resource(networkAttatchmentDefinition).
		Namespace(ns.Name).
		Create(makeNetworkAttachmentDefinition(networkAttatchmentConfigName, ns.Name), metav1.CreateOptions{})
	if err != nil {
		log.Infof("NetworkAttachmentDef create error: %s %v", ns.Name, err)
	}
}

func (c *Controller) onNamespaceDelete(obj interface{}) {
	ns, ok := obj.(*corev1.Namespace)
	if !ok {
		return
	}
	log.Infof("Namespace delete: %s %s", ns.Namespace, ns.Name)

	err := c.kubeDynamicClientset.
		Resource(networkAttatchmentDefinition).
		Namespace(ns.Name).
		Delete(networkAttatchmentConfigName, &metav1.DeleteOptions{})
	if err != nil {
		log.Infof("NetworkAttachmentDef delete error: %s %v", ns.Name, err)
	}
}
