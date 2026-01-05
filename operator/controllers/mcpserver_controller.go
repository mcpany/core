// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mcpv1alpha1 "github.com/mcpany/core/operator/api/v1alpha1"
)

// MCPServerReconciler reconciles a MCPServer object
type MCPServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mcp.any,resources=mcpservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mcp.any,resources=mcpservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mcp.any,resources=mcpservers/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *MCPServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Fetch the MCPServer instance
	mcpServer := &mcpv1alpha1.MCPServer{}
	err := r.Get(ctx, req.NamespacedName, mcpServer)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// 1. Check if the Deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: mcpServer.Name, Namespace: mcpServer.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForMCPServer(mcpServer)
		err = r.Create(ctx, dep)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// 2. Ensure the deployment size is the same as the spec
	size := mcpServer.Spec.Replicas
	if *found.Spec.Replicas != *size {
		found.Spec.Replicas = size
		err = r.Update(ctx, found)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// 3. Check if the Service already exists, if not create a new one
	foundService := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: mcpServer.Name, Namespace: mcpServer.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		// Define a new service
		svc := r.serviceForMCPServer(mcpServer)
		err = r.Create(ctx, svc)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// 4. Update the MCPServer status with the pod names
	// List the pods for this mcpServer's deployment
	// podList := &corev1.PodList{}
	// listOpts := []client.ListOption{
	// 	client.InNamespace(mcpServer.Namespace),
	// 	client.MatchingLabels(labelsForMCPServer(mcpServer.Name)),
	// }
	// if err = r.List(ctx, podList, listOpts...); err != nil {
	// 	return ctrl.Result{}, err
	// }
	// podNames := getPodNames(podList.Items)

	// Update status.Status if needed (simple check)
	if mcpServer.Status.AvailableReplicas != found.Status.AvailableReplicas {
		mcpServer.Status.AvailableReplicas = found.Status.AvailableReplicas
		err := r.Status().Update(ctx, mcpServer)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// deploymentForMCPServer returns a mcpServer Deployment object
func (r *MCPServerReconciler) deploymentForMCPServer(m *mcpv1alpha1.MCPServer) *appsv1.Deployment {
	ls := labelsForMCPServer(m.Name)
	replicas := m.Spec.Replicas

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   m.Spec.Image,
						Name:    "mcp-server",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "http",
						}},
						Args: []string{"run", "--config-path", "/etc/mcp/config.yaml"},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "config-volume",
							MountPath: "/etc/mcp",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "config-volume",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: m.Spec.ConfigMap,
								},
							},
						},
					}},
				},
			},
		},
	}
	// Set MCPServer instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// serviceForMCPServer returns a mcpServer Service object
func (r *MCPServerReconciler) serviceForMCPServer(m *mcpv1alpha1.MCPServer) *corev1.Service {
	ls := labelsForMCPServer(m.Name)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{{
				Port: 8080,
				Name: "http",
			}},
			Type: corev1.ServiceType(m.Spec.ServiceType),
		},
	}
	// Set MCPServer instance as the owner and controller
	ctrl.SetControllerReference(m, svc, r.Scheme)
	return svc
}

// labelsForMCPServer returns the labels for selecting the resources
// belonging to the given mcpServer CR name.
func labelsForMCPServer(name string) map[string]string {
	return map[string]string{"app": "mcp-server", "mcp_cr": name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *MCPServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpv1alpha1.MCPServer{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
