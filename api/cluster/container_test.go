// Copyright 2020 The Merlin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build unit

package cluster

import (
	"testing"

	fakeserving "github.com/kubeflow/kfserving/pkg/client/clientset/versioned/fake"
	fakeservingv1alpha2 "github.com/kubeflow/kfserving/pkg/client/clientset/versioned/typed/serving/v1alpha2/fake"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	ktesting "k8s.io/client-go/testing"

	"github.com/gojek/merlin/config"
)

func TestContainer_GetContainers(t *testing.T) {

	type args struct {
		namespace     string
		labelSelector string
	}
	tests := []struct {
		args
		mock      *v1.PodList
		wantError bool
	}{
		{
			args{
				labelSelector: "serving.kubeflow.org/inferenceservice=my-service",
				namespace:     "my-namespace",
			},
			&v1.PodList{
				Items: []v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "pod-1",
							Labels: map[string]string{
								"serving.kubeflow.org/inferenceservice": "my-service",
							},
						},
						Spec: v1.PodSpec{
							InitContainers: []v1.Container{
								{
									Name: "init-container-0",
								},
							},
							Containers: []v1.Container{
								{
									Name: "container-0",
								},
							},
						},
					},
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		kfClient := fakeserving.NewSimpleClientset().ServingV1alpha2().(*fakeservingv1alpha2.FakeServingV1alpha2)
		v1Client := fake.NewSimpleClientset().CoreV1()
		fakePodCtl := v1Client.Pods(tt.args.namespace).(*fakecorev1.FakePods)
		fakePodCtl.Fake.PrependReactor(listMethod, podResource, func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, tt.mock, nil
		})

		clusterMetadata := Metadata{GcpProject: "my-gcp", ClusterName: "my-cluster"}

		containerFetcher := NewContainerFetcher(v1Client, clusterMetadata)
		ctl, _ := newController(kfClient, v1Client, config.DeploymentConfig{}, containerFetcher)
		containers, err := ctl.GetContainers(tt.args.namespace, tt.args.labelSelector)
		if !tt.wantError {
			assert.NoErrorf(t, err, "expected no error got %v", err)
		} else {
			assert.Errorf(t, err, "expected error")
			return
		}

		assert.NotNil(t, containers)
		for _, pod := range tt.mock.Items {
			for _, initContainer := range pod.Spec.InitContainers {
				assert.True(t, isIn(initContainer, containers, pod.Name))
			}

			for _, container := range pod.Spec.Containers {
				assert.True(t, isIn(container, containers, pod.Name))
			}
		}
	}
}