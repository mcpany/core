// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package controllers provides the controller logic for the MCP Operator.
package controllers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	mcpv1alpha1 "github.com/mcpany/core/k8s/operator/api/v1alpha1"
)

func TestMCPServerReconciler_Reconcile(t *testing.T) {
	// Setup
	mcpServer := &mcpv1alpha1.MCPServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-mcp",
			Namespace: "default",
		},
		Spec: mcpv1alpha1.MCPServerSpec{
			Image:    "mcp-any:latest",
			Replicas: func(i int32) *int32 { return &i }(1),
		},
	}

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = mcpv1alpha1.AddToScheme(scheme)

	objs := []runtime.Object{mcpServer}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...).Build()

	r := &MCPServerReconciler{Client: cl, Scheme: scheme}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-mcp",
			Namespace: "default",
		},
	}

	// Execute
	res, err := r.Reconcile(context.Background(), req)

	// Verify
	assert.NoError(t, err)
	assert.True(t, res.Requeue) // Should requeue because it creates a deployment
}
