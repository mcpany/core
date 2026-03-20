import re

with open("k8s/operator/api/v1alpha1/zz_generated.deepcopy.go", "r") as f:
    content = f.read()

# Replace params without types to having types
content = content.replace(
"""//   - in: The input object.
//   - out: The output object.""",
"""//   - in (any): The input object.
//   - out (any): The output object."""
)

with open("k8s/operator/api/v1alpha1/zz_generated.deepcopy.go", "w") as f:
    f.write(content)
