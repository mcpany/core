// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ToolSpec defines the desired state of Tool.
type ToolSpec struct {
	// Name is the unique name of the tool.
	Name string `json:"name" validate:"required"`
	// Description is a human-readable description of what the tool does.
	Description string `json:"description"`
	// Command is the command used to execute the tool.
	Command []string `json:"command"`
	// Args are the arguments passed to the tool command.
	Args []string `json:"args"`
}

// ToolStatus defines the observed state of Tool.
type ToolStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Tool is the Schema for the tools API.
type Tool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ToolSpec   `json:"spec,omitempty"`
	Status ToolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ToolList contains a list of Tool.
type ToolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tool{}, &ToolList{})
}
