# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

from playwright.sync_api import sync_playwright
import time

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    page = browser.new_page()

    # Mock Services API
    page.route("**/api/v1/services", lambda route: route.fulfill(
        status=200,
        content_type="application/json",
        body='''[
            {
                "name": "weather-service",
                "id": "weather-service",
                "version": "1.0.0",
                "disable": false,
                "httpService": {"address": "http://wttr.in"},
                "tags": ["external", "weather"]
            },
            {
                "name": "db-service",
                "id": "db-service",
                "version": "2.1.0",
                "disable": false,
                "grpcService": {"address": "localhost:50051"},
                "tags": ["internal", "database"]
            }
        ]'''
    ))

    # Mock Topology API (for metrics)
    # We send multiple responses to simulate history if needed, but context accumulates history.
    # We just need one successful fetch to show a point.
    page.route("**/api/v1/topology", lambda route: route.fulfill(
        status=200,
        content_type="application/json",
        body='''{
            "core": {
                "id": "core",
                "type": "NODE_TYPE_CORE",
                "children": [
                    {
                        "id": "weather-service",
                        "type": "NODE_TYPE_SERVICE",
                        "status": "NODE_STATUS_ACTIVE",
                        "metrics": {"latencyMs": 120, "errorRate": 0, "qps": 5}
                    },
                    {
                        "id": "db-service",
                        "type": "NODE_TYPE_SERVICE",
                        "status": "NODE_STATUS_ACTIVE",
                        "metrics": {"latencyMs": 45, "errorRate": 0, "qps": 20}
                    }
                ]
            }
        }'''
    ))

    # Mock Doctor for System Status Banner to avoid error
    page.route("**/api/v1/doctor", lambda route: route.fulfill(
        status=200,
        content_type="application/json",
        body='''{"status": "ok", "timestamp": "2025-05-20T12:00:00Z", "checks": {}}'''
    ))

    # Mock settings
    page.route("**/api/v1/settings", lambda route: route.fulfill(
        status=200,
        content_type="application/json",
        body='''{}'''
    ))

    # Mock user
    page.route("**/api/v1/user/me", lambda route: route.fulfill(
        status=200,
        content_type="application/json",
        body='''{"id": "admin", "role": "admin"}'''
    ))

    # Navigate to Services page
    print("Navigating to Services page...")
    try:
        page.goto("http://localhost:9002/services", timeout=30000)
    except Exception as e:
        print(f"Navigation failed: {e}")
        # Capture screenshot anyway to see error
        page.screenshot(path="verification/error.png")
        browser.close()
        return

    # Wait for table to load
    print("Waiting for table...")
    try:
        page.wait_for_selector("table", timeout=10000)
    except:
        print("Table not found")

    # Wait a bit for Sparkline to render (it depends on topology fetch)
    print("Waiting for sparklines...")
    page.wait_for_timeout(3000)

    # Screenshot
    print("Taking screenshot...")
    page.screenshot(path="verification/sparklines.png")

    # Save to audit folder as requested
    import datetime
    date_str = datetime.date.today().strftime("%Y-%m-%d")
    page.screenshot(path=f".audit/ui/{date_str}/service_health_sparklines.png")

    browser.close()

with sync_playwright() as playwright:
    run(playwright)
