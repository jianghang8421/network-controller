package controller

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	EventMacvlanSubnetError = "MacvlanSubnetError"
	EventMacvlanIPError     = "MacvlanIPError"

	MessageNoEnoughIP = "No enough ip resouce in subnet: %s"
)

func (c *Controller) eventMacvlanSubnetError(pod *corev1.Pod, err error) {
	c.recorder.Event(pod, corev1.EventTypeNormal, EventMacvlanSubnetError, err.Error())
}

func (c *Controller) eventMacvlanIPError(pod *corev1.Pod, err error) {
	c.recorder.Event(pod, corev1.EventTypeNormal, EventMacvlanIPError, err.Error())
}
