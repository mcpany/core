# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import time
import requests
import json
from playwright.sync_api import sync_playwright

SERVICE_NAME = "verify-image-service"
BACKEND_URL = "http://localhost:50050"
FRONTEND_URL = "http://localhost:3000"

def register_service():
    url = f"{BACKEND_URL}/api/v1/services"
    payload = {
        "name": SERVICE_NAME,
        "command_line_service": {
            "command": "echo",
            "tools": [
                { "name": "get_image", "call_id": "call1", "description": "Returns an image" }
            ],
            "calls": {
                "call1": {
                    "args": [
                        json.dumps([
                            {
                                "type": "image",
                                "data": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=",
                                "mimeType": "image/png"
                            }
                        ])
                    ]
                }
            }
        }
    }
    # Clean up first
    try:
        requests.delete(f"{url}/{SERVICE_NAME}")
    except:
        pass

    resp = requests.post(url, json=payload)
    if not resp.ok:
        print(f"Failed to register service: {resp.text}")
        exit(1)
    print("Service registered.")

def run_verification():
    # Verify API lists the tool
    time.sleep(2)
    try:
        tools_resp = requests.get(f"{BACKEND_URL}/api/v1/tools")
        print("Tools API Response:", tools_resp.status_code)
        tools = tools_resp.json().get("tools", [])
        tool_names = [t["name"] for t in tools]
        print("Available tools:", tool_names)

        target_tool = f"{SERVICE_NAME}.get_image"
        if target_tool not in tool_names:
            print(f"ERROR: {target_tool} not found in API list.")
            # exit(1) # Continue to see screenshot
    except Exception as e:
        print("Failed to fetch tools:", e)

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        print(f"Navigating to {FRONTEND_URL}/playground...")
        page.goto(f"{FRONTEND_URL}/playground")

        # Wait for service/tool
        print(f"Waiting for {SERVICE_NAME}.get_image...")
        try:
            page.get_by_text(f"{SERVICE_NAME}.get_image").click(timeout=10000)
        except:
            print("Tool not found, refreshing...")
            page.reload()
            page.get_by_text(f"{SERVICE_NAME}.get_image").click(timeout=30000)

        print("Building command...")
        page.get_by_role("button", name="Build Command").click()

        # Wait for dialog close
        # page.wait_for_selector('role=dialog', state='hidden')
        time.sleep(1)

        print("Sending command...")
        page.get_by_label("Send").click()

        print("Waiting for image...")
        # Check for image with specific src prefix
        src_prefix = "data:image/png;base64,iVBOR"
        page.locator(f'img[src^="{src_prefix}"]').wait_for(state="visible", timeout=15000)

        print("Taking screenshot...")
        page.screenshot(path="verification/verification.png")
        browser.close()

if __name__ == "__main__":
    register_service()
    run_verification()
