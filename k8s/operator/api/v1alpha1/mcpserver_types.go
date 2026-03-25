// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MCPServerSpec defines the desired state of MCPServer.
type MCPServerSpec struct {
	// Image is the Docker image to run for the MCP Server.
	Image string `json:"image" validate:"required"`
	// Replicas is the number of desired instances of the MCP Server.
	Replicas *int32 `json:"replicas"`
	// ConfigMap is the name of the ConfigMap containing the MCP configuration.
	ConfigMap string `json:"configMap"`
	// ServiceType is the Kubernetes ServiceType for the MCP Server.
	ServiceType string `json:"serviceType"`
}

// MCPServerStatus defines the observed state of MCPServer.
type MCPServerStatus struct {
	// AvailableReplicas is the number of available instances of the MCP Server.
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
