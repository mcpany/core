// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0.

package controllers

import (
	"context"
	"testing"

	mcpv1alpha1 "github.com/mcpany/core/operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestMCPServerReconciler_Reconcile(t *testing.T) {
	s := scheme.Scheme
	if err := mcpv1alpha1.SchemeBuilder.AddToScheme(s); err != nil {
		t.Fatalf("failed to add to scheme: %v", err)
	}

	replicas := int32(2)

	mcpServer := &mcpv1alpha1.MCPServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-mcp-server",
			Namespace: "default",
		},
		Spec: mcpv1alpha1.MCPServerSpec{
			Replicas:    &replicas,
			Image:       "mcpany/server:latest",
			ServiceType: "ClusterIP",
			ConfigMap:   "mcp-config",
		},
	}

	cl := fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(mcpServer).Build()

	r := &MCPServerReconciler{
		Client: cl,
		Scheme: s,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-mcp-server",
			Namespace: "default",
		},
	}

	res, err := r.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected (Deployment creation)")
	}

	res, err = r.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("reconcile 2: (%v)", err)
	}

	found := &appsv1.Deployment{}
	err = cl.Get(context.Background(), types.NamespacedName{Name: "test-mcp-server", Namespace: "default"}, found)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}

	if *found.Spec.Replicas != replicas {
		t.Errorf("expected replicas %d, got %d", replicas, *found.Spec.Replicas)
	}

	if len(found.Spec.Template.Spec.Containers) == 0 {
		t.Fatal("no containers found in deployment")
	}
	container := found.Spec.Template.Spec.Containers[0]
	if container.Image != "mcpany/server:latest" {
		t.Errorf("expected image mcpany/server:latest, got %s", container.Image)
	}
}
