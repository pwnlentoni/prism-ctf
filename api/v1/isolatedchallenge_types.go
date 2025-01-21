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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IsolatedChallengeSpec defines the desired state of IsolatedChallenge.
type IsolatedChallengeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Containers []ContainerSpec `json:"containers"`
	Exposes    []ExposeSpec    `json:"exposes"`
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=date-time
	AvailableAt *metav1.Time `json:"available_at,omitempty"`
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	Lifetime *metav1.Duration `json:"lifetime,omitempty"`
}

// IsolatedChallengeStatus defines the observed state of IsolatedChallenge.
type IsolatedChallengeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// IsolatedChallenge is the Schema for the isolatedchallenges API.
type IsolatedChallenge struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IsolatedChallengeSpec   `json:"spec,omitempty"`
	Status IsolatedChallengeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IsolatedChallengeList contains a list of IsolatedChallenge.
type IsolatedChallengeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IsolatedChallenge `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IsolatedChallenge{}, &IsolatedChallengeList{})
}
