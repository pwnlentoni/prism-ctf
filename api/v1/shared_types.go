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

// +kubebuilder:validation:Enum=TCP;HTTP;UDP
type ExposeProtocol string

const (
	ExposeProtocolTCP  ExposeProtocol = "TCP"
	ExposeProtocolHTTP ExposeProtocol = "HTTP"
	ExposeProtocolUDP  ExposeProtocol = "UDP"
)

type ExposeSpec struct {
	Name      string `json:"name,omitempty"`
	Container string `json:"container"`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	Port     int            `json:"port"`
	Protocol ExposeProtocol `json:"protocol"`
}

type ExposeStatus struct {
	Name     string         `json:"name,omitempty"`
	Hostname string         `json:"hostname"`
	Protocol ExposeProtocol `json:"protocol"`
	Port     int            `json:"port,omitempty"`
}
