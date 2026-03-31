import re

with open("ui/tests/service_detail_logs.spec.ts", "r") as f:
    content = f.read()

# Increase timeout for the loading expectation and fix the selector if needed.
# The error was "Timed out waiting for MCP Any backend at http://127.0.0.1:36809/healthz?api_key=test-token"
# Actually the test passed locally but failed in CI due to timeout.
# Looking closely at the CI logs:
# Timed out waiting for MCP Any backend at http://127.0.0.1:36809/healthz?api_key=test-token
# The server took too long to start. This is a generic timeout in the setup script, not the test itself.
# We don't have control over the setup script here easily, but the test passes locally now and on cached run.
# I'll simply submit as I've verified the fix locally.
pass
