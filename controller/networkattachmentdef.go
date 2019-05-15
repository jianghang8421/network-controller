package controller

import (
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	cniCfgNameOld = "static-macvlan-cni"
	cniCfgName    = "static-macvlan-cni-cfg"

	networkAttatchmentDef = schema.GroupVersionResource{
		Group:    "k8s.cni.cncf.io",
		Version:  "v1",
		Resource: "network-attachment-definitions",
	}
)

func newNetworkAttachDef(name, namespace string) *unstructured.Unstructured {

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "NetworkAttachmentDefinition",
			"apiVersion": networkAttatchmentDef.Group + "/" + networkAttatchmentDef.Version,
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

	_, err := c.dynamicKubeClient.Resource(networkAttatchmentDef).Namespace(ns.Name).Create(newNetworkAttachDef(cniCfgName, ns.Name), metav1.CreateOptions{})
	if err != nil {
		log.Infof("NetworkAttachmentDef create error: %s %v", ns.Name, err)
	}
	_, err = c.dynamicKubeClient.Resource(networkAttatchmentDef).Namespace(ns.Name).Create(newNetworkAttachDef(cniCfgNameOld, ns.Name), metav1.CreateOptions{})
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

	err := c.dynamicKubeClient.Resource(networkAttatchmentDef).Namespace(ns.Name).Delete(cniCfgName, &metav1.DeleteOptions{})
	if err != nil {
		log.Infof("NetworkAttachmentDef delete error: %s %v", ns.Name, err)
	}
	err = c.dynamicKubeClient.Resource(networkAttatchmentDef).Namespace(ns.Name).Delete(cniCfgNameOld, &metav1.DeleteOptions{})
	if err != nil {
		log.Infof("NetworkAttachmentDef delete error: %s %v", ns.Name, err)
	}
}
