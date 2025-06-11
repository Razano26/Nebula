/*
Copyright 2025.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StunnerIngressSpec defines the desired state of StunnerIngress
type StunnerIngressSpec struct {
	// Target specifies the service to expose through the stunner ingress
	// +kubebuilder:validation:Required
	Target TargetRef `json:"target"`

	// Port specifies the port to be exposed through stunner
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`

	// Protocol specifies the protocol to be used (UDP/TCP)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=UDP;TCP
	Protocol string `json:"protocol"`

	// ExternalPort specifies the port on which the service will be exposed externally
	// If not specified, the same port as Target port will be used
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	ExternalPort *int32 `json:"externalPort,omitempty"`
}

// TargetRef defines the target service to expose
type TargetRef struct {
	// Name of the service to expose
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace of the service
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// StunnerIngressStatus defines the observed state of StunnerIngress
type StunnerIngressStatus struct {
	// ExternalAddresses contains the list of addresses where the service is exposed
	// +optional
	ExternalAddresses []string `json:"externalAddresses,omitempty"`

	// Conditions represent the latest available observations of the StunnerIngress state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// StunnerIngress is the Schema for the stunneringresses API
type StunnerIngress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StunnerIngressSpec   `json:"spec,omitempty"`
	Status StunnerIngressStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StunnerIngressList contains a list of StunnerIngress
type StunnerIngressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StunnerIngress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StunnerIngress{}, &StunnerIngressList{})
}
