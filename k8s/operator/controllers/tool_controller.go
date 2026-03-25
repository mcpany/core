// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package controllers provides the controller logic for the MCP Operator.
package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mcpanyv1alpha1 "github.com/mcpany/core/k8s/operator/api/v1alpha1"
)

// ToolReconciler reconciles a Tool object.
type ToolReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// +kubebuilder:rbac:groups=mcpany.mcp.so,resources=tools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mcpany.mcp.so,resources=tools/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mcpany.mcp.so,resources=tools/finalizers,verbs=update
func (r *ToolReconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ToolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpanyv1alpha1.Tool{}).
		Complete(r)
}
