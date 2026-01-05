// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
	// Register operator types with the scheme
	s := scheme.Scheme
	_ = mcpv1alpha1.SchemeBuilder.AddToScheme(s)

	replicas := int32(2)

	// Create a dummy MCPServer object
	mcpServer := &mcpv1alpha1.MCPServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-mcp-server",
			Namespace: "default",
		},
		Spec: mcpv1alpha1.MCPServerSpec{
			Replicas:    &replicas,
			Image:       "mcpany/server:latest",
			ServiceType: "ClusterIP",
			ConfigMap:   "my-config-map",
		},
	}

	// Create a fake client to mock API calls
	cl := fake.NewClientBuilder().WithScheme(s).WithRuntimeObjects(mcpServer).Build()

	// Create a ReconcileMCPServer object with the scheme and fake client
	r := &MCPServerReconciler{
		Client: cl,
		Scheme: s,
	}

	// Mock request to simulate Reconcile() being called on an event for a watched resource
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-mcp-server",
			Namespace: "default",
		},
	}

	// Execute Reconcile
	res, err := r.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation
	//nolint:staticcheck // Requeue is deprecated but we still check it
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected (Deployment creation)")
	}

	// Check if Deployment was created
	found := &appsv1.Deployment{}
	err = cl.Get(context.Background(), types.NamespacedName{Name: "test-mcp-server", Namespace: "default"}, found)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}

	// Verify Deployment Spec
	if *found.Spec.Replicas != replicas {
		t.Errorf("expected replicas %d, got %d", replicas, *found.Spec.Replicas)
	}

	container := found.Spec.Template.Spec.Containers[0]
	if container.Image != "mcpany/server:latest" {
		t.Errorf("expected image mcpany/server:latest, got %s", container.Image)
	}

	// Verify Volume Mounts
	foundVolumeMount := false
	for _, vm := range container.VolumeMounts {
		if vm.Name == "config-volume" && vm.MountPath == "/etc/mcp" {
			foundVolumeMount = true
			break
		}
	}
	if !foundVolumeMount {
		t.Error("expected config-volume mount at /etc/mcp")
	}

	// Verify Arguments
	foundArgs := false
	for i, arg := range container.Args {
		if arg == "--config-path" && i+1 < len(container.Args) && container.Args[i+1] == "/etc/mcp/config.yaml" {
			foundArgs = true
			break
		}
	}
	if !foundArgs {
		t.Error("expected --config-path /etc/mcp/config.yaml argument")
	}

	// Verify Volumes
	foundVolume := false
	for _, v := range found.Spec.Template.Spec.Volumes {
		//nolint:staticcheck // We can't remove embedded field VolumeSource
		if v.Name == "config-volume" && v.VolumeSource.ConfigMap != nil && v.VolumeSource.ConfigMap.Name == "my-config-map" {
			foundVolume = true
			break
		}
	}
	if !foundVolume {
		t.Error("expected config-volume from ConfigMap my-config-map")
	}
}
