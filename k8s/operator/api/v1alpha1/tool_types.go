// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package v1alpha1 contains API Schema definitions for the v1alpha1 API group.
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ToolSpec defines the desired state of Tool.
//
// Summary: Desired configuration for a Tool instance.
type ToolSpec struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Command     []string `json:"command"`
	Args        []string `json:"args"`
}

// ToolStatus defines the observed state of Tool.
//
// Summary: Real-time observed status of a Tool instance.
type ToolStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Tool is the Schema for the tools API.
//
// Summary: Top-level custom resource object representing a Tool.
type Tool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ToolSpec   `json:"spec,omitempty"`
	Status ToolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ToolList contains a list of Tool.
//
// Summary: List of Tool resources.
type ToolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tool{}, &ToolList{})
}
