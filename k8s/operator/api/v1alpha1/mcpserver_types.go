// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MCPServerSpec mCPServerSpec represents a mcp server spec.
//
// Summary: MCPServerSpec represents a mcp server spec.
type MCPServerSpec struct {
	// Replicas is the number of replicas for the server
	Replicas *int32 `json:"replicas,omitempty"`
	// Image is the container image to use
	Image string `json:"image,omitempty"`
	// ServiceType is the type of Kubernetes Service to expose (ClusterIP, LoadBalancer, NodePort)
	ServiceType string `json:"serviceType,omitempty"`
	// ConfigMap is the name of the ConfigMap containing config.yaml
	// +kubebuilder:validation:Required
	ConfigMap string `json:"configMap"`
}

// MCPServerStatus mCPServerStatus represents a mcp server status.
//
// Summary: MCPServerStatus represents a mcp server status.
type MCPServerStatus struct {
	// AvailableReplicas is the number of available replicas
	AvailableReplicas int32 `json:"availableReplicas"`
}


// MCPServer mCPServer represents a mcp server.
//
// Summary: MCPServer represents a mcp server.
type MCPServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MCPServerSpec   `json:"spec,omitempty"`
	Status MCPServerStatus `json:"status,omitempty"`
}


// MCPServerList mCPServerList represents a mcp server list.
//
// Summary: MCPServerList represents a mcp server list.
type MCPServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MCPServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MCPServer{}, &MCPServerList{})
}
