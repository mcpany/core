import re

with open("k8s/operator/controllers/mcpserver_controller.go", "r") as f:
    content = f.read()

content = content.replace(
"""//   - m: The MCPServer resource.""",
"""//   - m (*mcpv1alpha1.MCPServer): The MCPServer resource."""
)

content = content.replace(
"""//   - name: The name of the MCPServer resource.""",
"""//   - name (string): The name of the MCPServer resource."""
)

with open("k8s/operator/controllers/mcpserver_controller.go", "w") as f:
    f.write(content)
