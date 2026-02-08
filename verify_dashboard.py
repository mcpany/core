from playwright.sync_api import Page, expect, sync_playwright
import time

def verify_dashboard():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Mock Service Config
        page.route("**/api/v1/services/test-service", lambda route: route.fulfill(
            status=200,
            content_type="application/json",
            body='{"service": {"id": "test-service", "name": "test-service", "version": "1.0.0", "httpService": {"address": "http://localhost:8080", "resources": [{"uri": "res://test", "mimeType": "text/plain"}]}, "tags": ["production"]}}'
        ))

        # Mock Service Status
        page.route("**/api/v1/services/test-service/status", lambda route: route.fulfill(
            status=200,
            content_type="application/json",
            body='{"tools": [{"name": "test-tool", "description": "A test tool", "inputSchema": "{}"}], "metrics": {"uptime": 3600}}'
        ))

        # Mock Topology (for overview sparklines via context)
        page.route("**/api/v1/topology", lambda route: route.fulfill(
            status=200,
            content_type="application/json",
            body='{"core": {"id": "core", "type": "NODE_TYPE_CORE", "children": [{"id": "test-service", "type": "NODE_TYPE_SERVICE", "metrics": {"latencyMs": 120, "errorRate": 0.05, "qps": 10}}]}}'
        ))

        print("Navigating to dashboard...")
        page.goto("http://localhost:9002/upstream-services/test-service")

        # Wait for loading to finish
        try:
            page.wait_for_selector("text=test-service", timeout=10000)
            page.wait_for_selector("text=Overview", timeout=5000)
        except Exception as e:
            print("Timeout waiting for selector")
            page.screenshot(path="verification_error.png")
            raise e

        # Click Tabs to load content
        page.click("text=Overview")
        time.sleep(1) # Wait for animation/render
        page.screenshot(path="verification_overview.png")

        page.click("text=Tools")
        time.sleep(1)
        page.screenshot(path="verification_tools.png")

        page.click("text=Resources")
        time.sleep(1)
        page.screenshot(path="verification_resources.png")

        print("Screenshots taken.")
        browser.close()

if __name__ == "__main__":
    verify_dashboard()
