// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package v1alpha1 contains API Schema definitions for the v1alpha1 API group.
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MCPServerSpec defines the desired state of MCPServer.
//
// Summary: Desired configuration for an MCPServer instance.
type MCPServerSpec struct {
	Image       string `json:"image"`
	Replicas    *int32 `json:"replicas"`
	ConfigMap   string `json:"configMap"`
	ServiceType string `json:"serviceType"`
}

// MCPServerStatus defines the observed state of MCPServer.
//
// Summary: Real-time observed status of an MCPServer instance.
type MCPServerStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MCPServer is the Schema for the mcpservers API.
//
// Summary: Top-level custom resource object representing an MCPServer.
type MCPServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MCPServerSpec   `json:"spec,omitempty"`
	Status MCPServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MCPServerList contains a list of MCPServer.
//
// Summary: List of MCPServer resources.
type MCPServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MCPServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MCPServer{}, &MCPServerList{})
}
