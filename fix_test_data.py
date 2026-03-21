import re

filepath = 'ui/tests/e2e/test-data.ts'
with open(filepath, 'r') as f:
    content = f.read()

# Add service-1, service-2, service-3 to the seedGlobalState
# Wait, let's just make the tests look for Payment Gateway and User Service OR
# we can just add service-1 and service-2 to test-data.ts

new_services = """
        {
            id: "service-1",
            name: "service-1",
            version: "v1.0",
            http_service: { address: "http://localhost:8080" }
        },
        {
            id: "service-2",
            name: "service-2",
            version: "v1.0",
            http_service: { address: "http://localhost:8080" }
        },
        {
            id: "service-3",
            name: "service-3",
            version: "v1.0",
            http_service: { address: "http://localhost:8080" }
        },
"""

content = content.replace('const services = [\n        {', 'const services = [\n' + new_services + '        {')

with open(filepath, 'w') as f:
    f.write(content)
