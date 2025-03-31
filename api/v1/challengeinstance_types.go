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

// ChallengeInstanceSpec defines the desired state of ChallengeInstance.
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.random_id) || has(self.random_id)", message="Random id is required once set"
// +kubebuilder:validation:XValidation:rule="!has(oldSelf.flag) || has(self.flag)", message="Flag is required once set"
type ChallengeInstanceSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Team is immutable"
	// +kubebuilder:validation:MaxLength=512
	Team string `json:"team"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Challenge is immutable"
	// +kubebuilder:validation:MaxLength=512
	Challenge string `json:"challenge"`
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=date-time
	Expiration *metav1.Time `json:"expiration,omitempty"`
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Random id is immutable"
	// +kubebuilder:validation:MaxLength=512
	RandomId string `json:"random_id"`
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Flag is immutable"
	// +kubebuilder:validation:MaxLength=512
	Flag string `json:"flag"`
}

// ChallengeInstanceStatus defines the observed state of ChallengeInstance.
type ChallengeInstanceStatus struct {
	ExposedUrls []ExposeStatus `json:"exposedUrls"`
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:selectablefield:JSONPath=`.spec.team`
// +kubebuilder:selectablefield:JSONPath=`.spec.challenge`

// ChallengeInstance is the Schema for the challengeinstances API.
type ChallengeInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChallengeInstanceSpec   `json:"spec,omitempty"`
	Status ChallengeInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ChallengeInstanceList contains a list of ChallengeInstance.
type ChallengeInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ChallengeInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ChallengeInstance{}, &ChallengeInstanceList{})
}
