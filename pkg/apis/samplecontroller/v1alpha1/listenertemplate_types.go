/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	pipelinev1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ListenerTemplate is a specification for a listenerTemplate resource
type ListenerTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ListenerTemplateSpec   `json:"spec"`
	Status ListenerTemplateStatus `json:"status"`
}

// ListenerTemplateSpec is the spec for a listenerTemplate resource
type ListenerTemplateSpec struct {
	Params      []TemplateParam                         `json:"params,omitempty"`
	Resources   []pipelinev1alpha1.PipelineResourceSpec `json:"resources,omitempty"`
	PipelineRun pipelinev1alpha1.PipelineRunSpec        `json:"pipelinerun,omitempty"`
}

// TemplateParam defines arbitrary parameters needed by a Pipelinerun, Resource defined in the ListenerTemplate.
type TemplateParam struct {
	Name string `json:"name"`
	// +optional
	Description string `json:"description,omitempty"`
	// +optional
	Default string `json:"default,omitempty"`
}

// ListenerTemplateStatus is the status for a ListenerTemplate resource
type ListenerTemplateStatus struct {
	AvailableReference int32 `json:"availableReference"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ListenerTemplateList is a list of ListenerTemplate resources
type ListenerTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ListenerTemplate `json:"items"`
}
