package controller

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	eventMacvlanSubnetError = "MacvlanSubnetError"
	eventMacvlanIPError     = "MacvlanIPError"

	messageNoEnoughIP = "No enough ip resouce in subnet: %s"
)

func (c *Controller) eventMacvlanSubnetError(pod *corev1.Pod, err error) {
	c.recorder.Event(pod, corev1.EventTypeNormal, eventMacvlanSubnetError, err.Error())
}

func (c *Controller) eventMacvlanIPError(pod *corev1.Pod, err error) {
	c.recorder.Event(pod, corev1.EventTypeNormal, eventMacvlanIPError, err.Error())
}
