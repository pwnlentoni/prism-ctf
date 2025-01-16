package v1

import (
	corev1 "k8s.io/api/core/v1"
)

type ContainerSpec struct {
	Ports  []PortSpec        `json:"ports"`
	Egress bool              `json:"egress"`
	Spec   *corev1.Container `json:"spec"`
	// +kubebuilder:default:=1
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas int `json:"replicas"`
}

type PortSpec struct {
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	Port     int             `json:"port"`
	Protocol corev1.Protocol `json:"proto"`
}

type ExposeSpec struct {
	Name      string `json:"name,omitempty"`
	Container string `json:"container"`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	Port int `json:"port"`
	// TODO: add UDP support
	// +kubebuilder:validation:Enum=TCP;HTTP;UDP
	Protocol string `json:"protocol"`
}
