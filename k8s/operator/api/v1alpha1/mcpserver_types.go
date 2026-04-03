// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MCPServerSpec represents the public MCPServerSpec entity.
//
// Summary: Defines the structured data model representing a server spec.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// MCPServerStatus represents the public MCPServerStatus entity.
//
// Summary: Defines the structured data model representing a server status.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type MCPServerStatus struct {
	// AvailableReplicas is the number of available replicas
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MCPServer represents the public MCPServer entity.
//
// Summary: Provides network listening and request routing capabilities for server.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type MCPServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MCPServerSpec   `json:"spec,omitempty"`
	Status MCPServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MCPServerList represents the public MCPServerList entity.
//
// Summary: Defines the structured data model representing a server list.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type MCPServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MCPServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MCPServer{}, &MCPServerList{})
}
