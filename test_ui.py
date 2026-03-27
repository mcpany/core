import re

file_path = "ui/src/components/diagnostics/connection-diagnostic.tsx"

with open(file_path, 'r') as f:
    content = f.read()

# Add the missing 'browser_connectivity' to the initial steps array if httpService is detected
search = """    if (service.websocketService || service.httpService) {
        initialSteps.splice(1, 0, { id: "browser_connectivity", name: "Browser Connectivity Check", status: "pending", logs: [] });
    }"""

# It looks like it actually IS present in line 58 of connection-diagnostic.tsx!
# Wait, let's verify if that line 58 is missing.
