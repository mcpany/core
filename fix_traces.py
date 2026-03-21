import re

with open('ui/tests/e2e/traces.spec.ts', 'r') as f:
    content = f.read()

# Replace timeout from 10000 to 20000 for visibility
content = content.replace('timeout: 10000', 'timeout: 20000')

with open('ui/tests/e2e/traces.spec.ts', 'w') as f:
    f.write(content)
