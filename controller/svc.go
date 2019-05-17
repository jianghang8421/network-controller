package controller

import corev1 "k8s.io/api/core/v1"

// SyncService do sync svc auto creating/deleting by pod
func (c *Controller) SyncService(pod *corev1.Pod) error {
	return nil
}
