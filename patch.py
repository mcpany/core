import re

with open("ui/tests/upstream_service_detail.spec.ts", "r") as f:
    content = f.read()

content = content.replace("`/service/${serviceName}`", "`/upstream-services/${serviceName}`")

with open("ui/tests/upstream_service_detail.spec.ts", "w") as f:
    f.write(content)
