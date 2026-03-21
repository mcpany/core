# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import urllib.request
import json
import sys

# Get the latest CircleCI workflow for this branch to see what's failing
req = urllib.request.Request("https://circleci.com/api/v2/project/github/mcpany/core/pipeline?branch=fix-daily-security-hardening")
try:
    with urllib.request.urlopen(req) as response:
        data = json.loads(response.read().decode())
        print(json.dumps(data, indent=2))
except Exception as e:
    print(f"Error fetching CircleCI data: {e}")
