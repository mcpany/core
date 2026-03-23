import re

with open('ui/tests/e2e/test-data.ts', 'r') as f:
    content = f.read()

# We need to type the objects to bypass any type errors since we removed the TS import
# Wait, let's see what playwright test output says.
