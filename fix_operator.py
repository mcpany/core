with open("k8s/operator/controllers/mcpserver_controller.go", "r") as f:
    content = f.read()

content = content.replace("""// MCPServerReconciler reconciles a MCPServer object
//
// Summary: Reconciles a MCPServer object in the cluster.
//
// Parameters:
//   - Client: The Kubernetes client.
//   - Scheme: The runtime scheme.
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - Modifies Kubernetes resources.""", "// MCPServerReconciler reconciles a MCPServer object\n//\n// Summary: Represents a reconciler for MCPServer objects.")

content = content.replace("""// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// It creates or updates the Deployment and Service for the MCPServer.
//
// Summary: Creates or updates Kubernetes resources for the MCPServer.
//
// Parameters:
//   - ctx: The context for the request.
//   - req: The reconciliation request containing the namespaced name of the MCPServer.
//
// Returns:
//   - ctrl.Result: The result of the reconciliation, indicating if the request should be requeued.
//   - error: Any error that occurred during reconciliation.
//
// Errors:
//   - Returns an error if fetching or updating resources fails.
//
// Side Effects:
//   - Creates or updates Deployments and Services in the cluster.""", """// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// It creates or updates the Deployment and Service for the MCPServer.
//
// Summary: Creates or updates Kubernetes resources for the MCPServer.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - req (ctrl.Request): The reconciliation request containing the namespaced name of the MCPServer.
//
// Returns:
//   - ctrl.Result: The result of the reconciliation, indicating if the request should be requeued.
//   - error: Any error that occurred during reconciliation.
//
// Errors:
//   - Returns an error if fetching or updating resources fails.
//
// Side Effects:
//   - Creates or updates Deployments and Services in the cluster.""")

content = content.replace("""// SetupWithManager sets up the controller with the Manager.
//
// Summary: Configures the controller with the provided manager.
//
// Parameters:
//   - mgr: The controller manager.
//
// Returns:
//   - error: Any error that occurred during setup.
//
// Errors:
//   - Returns an error if the controller cannot be set up.
//
// Side Effects:
//   - Registers the controller with the manager.""", """// SetupWithManager sets up the controller with the Manager.
//
// Summary: Configures the controller with the provided manager.
//
// Parameters:
//   - mgr (ctrl.Manager): The controller manager.
//
// Returns:
//   - error: Any error that occurred during setup.
//
// Errors:
//   - Returns an error if the controller cannot be set up.
//
// Side Effects:
//   - Registers the controller with the manager.""")


with open("k8s/operator/controllers/mcpserver_controller.go", "w") as f:
    f.write(content)
