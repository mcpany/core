# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import requests
import time
from playwright.sync_api import sync_playwright

BASE_URL = "http://localhost:9002"

def seed_traffic():
    print("Seeding traffic data...")
    # Attempt to seed traffic via API (might need auth)
    url = f"{BASE_URL}/api/v1/debug/seed_traffic"
    payload = [
        {"time": "12:00", "requests": 100, "errors": 2, "latency": 50},
        {"time": "12:01", "requests": 150, "errors": 0, "latency": 45}
    ]
    try:
        # Try with common dev key or no key
        res = requests.post(url, json=payload, headers={"X-API-Key": "mcp-admin-key"}, timeout=5)
        print(f"Seed response: {res.status_code}")
    except Exception as e:
        print(f"Seed request failed (server might not be up or unreachable): {e}")

def verify_visualizer():
    with sync_playwright() as p:
        print("Launching browser...")
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Login
        print("Logging in...")
        try:
            page.goto(f"{BASE_URL}/login", wait_until="networkidle", timeout=10000)
        except Exception as e:
            print(f"Failed to load login page: {e}")
            browser.close()
            return

        page.wait_for_timeout(1000)

        # Check if we are already logged in (redirected)
        if "/login" in page.url:
            page.fill('input[name="username"]', 'py-admin')
            page.fill('input[name="password"]', 'password')
            page.click('button[type="submit"]')
            try:
                page.wait_for_url(f"{BASE_URL}/", timeout=15000)
            except:
                print("Login redirect timeout, checking if we are dashboard anyway...")

        print("Logged in.")

        # Go to Visualizer
        print("Navigating to Visualizer...")
        page.goto(f"{BASE_URL}/visualizer", wait_until="networkidle")

        # Check for error toast
        if page.is_visible("text=Topology Error"):
            print("Warning: Topology Error toast detected.")

        # Wait for canvas or nodes
        print("Waiting for graph nodes...")
        try:
            # Wait up to 10s for nodes
            page.wait_for_selector('.react-flow__node', timeout=10000)
            nodes = page.locator('.react-flow__node')
            count = nodes.count()
            print(f"Found {count} nodes.")

            # Identify node types
            for i in range(count):
                node_text = nodes.nth(i).inner_text()
                print(f"Node {i}: {node_text}")

            if count > 0:
                print("✅ Visualizer verification PASSED: Nodes are visible.")
            else:
                print("❌ Visualizer verification FAILED: No nodes found.")

            page.screenshot(path="verification/visualizer.png")

        except Exception as e:
            print(f"❌ Visualizer verification FAILED (Timeout/Error): {e}")
            page.screenshot(path="verification/visualizer_error.png")

        browser.close()

if __name__ == "__main__":
    seed_traffic()
    verify_visualizer()
