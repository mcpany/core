import re

with open("k8s/operator/controllers/mcpserver_controller.go", "r") as f:
    content = f.read()

# Add missing Summary to deploymentForMCPServer
content = content.replace(
"""// deploymentForMCPServer creates a new Deployment for the MCPServer resource.
//
// Parameters:""",
"""// deploymentForMCPServer creates a new Deployment for the MCPServer resource.
//
// Summary: Creates a new Deployment for the MCPServer resource.
//
// Parameters:"""
)

# Add missing Throws/Errors and Side Effects to deploymentForMCPServer
content = content.replace(
"""// Returns:
//   - *appsv1.Deployment: The created Deployment.
func (r *MCPServerReconciler) deploymentForMCPServer""",
"""// Returns:
//   - *appsv1.Deployment: The created Deployment.
//
// Errors:
//   - Returns nil if setting controller reference fails.
//
// Side Effects:
//   - Creates a Deployment struct but does not apply it to the cluster directly.
func (r *MCPServerReconciler) deploymentForMCPServer"""
)


# Add missing Summary to serviceForMCPServer
content = content.replace(
"""// serviceForMCPServer creates a new Service for the MCPServer resource.
//
// Parameters:""",
"""// serviceForMCPServer creates a new Service for the MCPServer resource.
//
// Summary: Creates a new Service for the MCPServer resource.
//
// Parameters:"""
)

# Add missing Throws/Errors and Side Effects to serviceForMCPServer
content = content.replace(
"""// Returns:
//   - *corev1.Service: The created Service.
func (r *MCPServerReconciler) serviceForMCPServer""",
"""// Returns:
//   - *corev1.Service: The created Service.
//
// Errors:
//   - Returns nil if setting controller reference fails.
//
// Side Effects:
//   - Creates a Service struct but does not apply it to the cluster directly.
func (r *MCPServerReconciler) serviceForMCPServer"""
)

# Add missing Summary to labelsForMCPServer
content = content.replace(
"""// labelsForMCPServer returns the labels for selecting the resources
// belonging to the given mcpServer CR name.
//
// Parameters:""",
"""// labelsForMCPServer returns the labels for selecting the resources
// belonging to the given mcpServer CR name.
//
// Summary: Returns labels for selecting resources by MCPServer CR name.
//
// Parameters:"""
)

# Add missing Throws/Errors and Side Effects to labelsForMCPServer
content = content.replace(
"""// Returns:
//   - map[string]string: A map of labels.
func labelsForMCPServer""",
"""// Returns:
//   - map[string]string: A map of labels.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func labelsForMCPServer"""
)

with open("k8s/operator/controllers/mcpserver_controller.go", "w") as f:
    f.write(content)
