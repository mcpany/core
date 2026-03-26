// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MCPServerSpec defines the desired state of MCPServer.
type MCPServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster.
	// Important: Run "make" to regenerate code after modifying this file.

	// Image is the container image to run for the MCP server.
	Image string `json:"image"`

	// Replicas is the number of instances of the server to run.
	Replicas *int32 `json:"replicas,omitempty"`

	// ConfigMap is the name of the ConfigMap containing the server configuration.
	ConfigMap string `json:"configMap"`

	// ServiceType defines the type of service to create (ClusterIP, NodePort, LoadBalancer).
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	// +kubebuilder:default=ClusterIP
	ServiceType string `json:"serviceType,omitempty"`
}

// MCPServerStatus defines the observed state of MCPServer.
type MCPServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster.
	// Important: Run "make" to regenerate code after modifying this file.

	// AvailableReplicas is the number of healthy replicas currently running.
	AvailableReplicas int32 `json:"availableReplicas"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MCPServer is the Schema for the mcpservers API.
type MCPServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MCPServerSpec   `json:"spec,omitempty"`
	Status MCPServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MCPServerList contains a list of MCPServer.
type MCPServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MCPServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MCPServer{}, &MCPServerList{})
}
