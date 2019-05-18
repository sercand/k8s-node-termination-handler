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
	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	systemNamespace = "kube-system"
)

type podEvictionHandler struct {
	client corev1.CoreV1Interface
	node   string
}

// List all pods on the node
// Evict all pods on the node not in kube-system namespace
// Return nil on success
func NewPodEvictionHandler(node string, client *client.Clientset) PodEvictionHandler {
	return &podEvictionHandler{
		client: client.CoreV1(),
		node:   node,
	}
}

func (p *podEvictionHandler) EvictPods(excludePods map[string]string) error {
	query := []fields.Selector{
		fields.OneTermEqualSelector("spec.nodeName", p.node),
		fields.ParseSelectorOrDie("metadata.namespace!=" + systemNamespace),
	}
	for k := range excludePods {
		query = append(query, fields.ParseSelectorOrDie("metadata.name!="+k))
	}
	gracePeriod := int64(0)
	deleteOptions := &metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod}

	loptions := metav1.ListOptions{FieldSelector: fields.AndSelectors(query...).String()}

	result := &v1.PodList{}
	err := p.client.RESTClient().Get().
		Namespace(metav1.NamespaceAll).
		Resource("pods").
		VersionedParams(&loptions, scheme.ParameterCodec).
		Do().
		Into(result)
	if err != nil {
		glog.V(2).Infof("Failed to list pods on node %s, got error=%v", p.node,err)
		return err
	}
	nss := map[string]bool{}
	for _, x := range result.Items {
		nss[x.Namespace] = true
	}
	for ns := range nss {
		err = p.client.RESTClient().Delete().
			Throttle(nil).
			Namespace(ns).
			Resource("pods").
			VersionedParams(&loptions, scheme.ParameterCodec).
			Body(deleteOptions).
			Do().
			Error()
		if err != nil {
			glog.V(2).Infof("Failed to remove pods on node %s of namespace %s error=%v", p.node, ns, err)
			return err
		}
	}
	glog.V(1).Infof("Successfully evicted all pods from node %q", p.node)
	return nil
}
