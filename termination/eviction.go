// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package termination

import (
	"time"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	client "k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	systemNamespace = "kube-system"
	eventReason     = "NodeTermination"
)

type podEvictionHandler struct {
	client               corev1.CoreV1Interface
	node                 string
	systemPodGracePeriod time.Duration
}

// List all pods on the node
// Evict all pods on the node not in kube-system namespace
// Return nil on success
func NewPodEvictionHandler(node string, client *client.Clientset, systemPodGracePeriod time.Duration) PodEvictionHandler {
	return &podEvictionHandler{
		client:               client.CoreV1(),
		node:                 node,
		systemPodGracePeriod: systemPodGracePeriod,
	}
}

func (p *podEvictionHandler) EvictPods(excludePods map[string]string, timeout time.Duration) error {
	options := metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector("spec.nodeName", p.node).String()}
	pods, err := p.client.Pods(metav1.NamespaceAll).List(options)
	if err != nil {
		glog.V(2).Infof("Failed to list pods - %v", err)
		return err
	}
	var regularPods []v1.Pod
	// Separate pods in kube-system namespace such that they can be evicted at the end.
	// This is especially helpful in scenarios like reclaiming logs prior to node termination.
	for _, pod := range pods.Items {
		if ns, exists := excludePods[pod.Name]; !exists || ns != pod.Namespace {
			if pod.Namespace != systemNamespace {
				regularPods = append(regularPods, pod)
			}
		}
	}
	// Evict regular pods first.
	gracePeriod := int64(30)
	deleteOptions := &metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod}
	if err := p.deletePods(regularPods, deleteOptions); err != nil {
		return err
	}
	glog.V(4).Infof("Successfully evicted all pods from node %q", p.node)
	return nil
}

func (p *podEvictionHandler) deletePods(pods []v1.Pod, deleteOptions *metav1.DeleteOptions) error {
	for _, pod := range pods {
		glog.V(4).Infof("About to delete pod %q in namespace %q", pod.Name, pod.Namespace)
		go p.client.Pods(pod.Namespace).Delete(pod.Name, deleteOptions)
	}
	return nil
}
