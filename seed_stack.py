# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import requests
import sys

BASE_URL = "http://localhost:50050"

def seed_stack():
    stack_name = "production-stack"
    collection = {
        "name": stack_name,
        "services": [
            {
                "name": "web-server",
                "id": "web-server",
                "command_line_service": {
                    "command": "python3 -m http.server 8080",
                    "working_directory": "."
                }
            },
            {
                "name": "worker",
                "id": "worker",
                "command_line_service": {
                    "command": "echo 'working'",
                    "working_directory": "."
                }
            }
        ]
    }

    print(f"Seeding stack: {stack_name} to {BASE_URL}/api/v1/collections")

    # Using POST /api/v1/collections as supported by backend
    try:
        response = requests.post(f"{BASE_URL}/api/v1/collections", json=collection)

        if response.status_code == 201:
            print("✅ Stack seeded successfully (Created).")
        elif response.status_code == 200:
             print("✅ Stack seeded successfully (OK).")
        elif response.status_code == 405:
             # Try PUT if POST not allowed on /collections? No, api.go says POST is allowed.
             # Try PUT on detail
             print("⚠️ POST failed (405). Trying PUT...")
             response = requests.put(f"{BASE_URL}/api/v1/collections/{stack_name}", json=collection)
             if response.status_code in [200, 201]:
                 print("✅ Stack seeded successfully via PUT.")
             else:
                 print(f"❌ Failed to seed stack via PUT: {response.status_code} {response.text}")
                 sys.exit(1)
        else:
            print(f"❌ Failed to seed stack: {response.status_code} {response.text}")
            # Try to see if it already exists
            if "already exists" in response.text:
                 print("ℹ️ Stack already exists.")
            else:
                 sys.exit(1)

    except Exception as e:
        print(f"❌ Connection failed: {e}")
        sys.exit(1)

if __name__ == "__main__":
    seed_stack()
