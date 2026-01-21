// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ToolSpec defines the desired state of Tool
type ToolSpec struct {
	// Type is the type of tool (e.g., "container", "binary", "script")
	// +kubebuilder:validation:Enum=container;binary;script
	Type string `json:"type"`

	// Image is the container image to use (for type "container")
	Image string `json:"image,omitempty"`

	// Command is the command to run (for type "binary" or "script")
	Command []string `json:"command,omitempty"`

	// Args are the arguments to pass to the command
	Args []string `json:"args,omitempty"`

	// Content is the inline content (for type "script")
	Content string `json:"content,omitempty"`
}

// ToolStatus defines the observed state of Tool
type ToolStatus struct {
	// Ready indicates if the tool is ready to be used
	Ready bool `json:"ready"`
	// Message provides details about the status
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Tool is the Schema for the tools API
type Tool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ToolSpec   `json:"spec,omitempty"`
	Status ToolStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ToolList contains a list of Tool
type ToolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tool{}, &ToolList{})
}
