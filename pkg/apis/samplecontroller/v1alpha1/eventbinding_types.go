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
	duckv1beta1 "github.com/knative/pkg/apis/duck/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EventBinding is a specification for a eventBinding resource
type EventBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EventBindingSpec   `json:"spec"`
	Status EventBindingStatus `json:"status"`
}

// EventBindingSpec is the spec for a listenerTemplate resource
type EventBindingSpec struct {
	TemplateRef TemplateRef `json:"templateRef"`
	Event       Event       `json:"event"`
	Params      []Param     `json:"params,omitempty"`
}

// Event use to define a cloud event.
type Event struct {
	// Class of Cloudevent
	Class string `json:"class,omitempty"`
	// Type of Cloudevent
	Type string `json:"type,omitempty"`
}

// TemplateRef can be used to refer to a specific instance of a ListenerTemplate.
type TemplateRef struct {
	// Name of the referent
	Name string `json:"name,omitempty"`
	// API version of the referent
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`
}

// Param declares a value to use for the Param called Name.
type Param struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// EventBindingStatus is the status for a ListenerTemplate resource
type EventBindingStatus struct {
	duckv1beta1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EventBindingList is a list of EventBindingList resources
type EventBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []EventBinding `json:"items"`
}

// // HasReference returns true if AvailableReference in Status is not 0 .
// func (lt *ListenerTemplate) HasReference() bool {
// 	return lt.Status.AvailableReference != 0
// }
