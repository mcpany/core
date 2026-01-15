import os
import time
from playwright.sync_api import sync_playwright

def verify_test_connection_button():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Mock Services List API
        page.route("**/api/v1/services", lambda route: route.fulfill(
            status=200,
            content_type="application/json",
            body='[{"name": "test-service", "disable": false, "version": "1.0.0", "mcp_service": {"stdio_connection": {"command": "echo"}}}]'
        ))

        # Mock Service Status API
        page.route("**/api/v1/services/test-service/status?check=true", lambda route: route.fulfill(
            status=200,
            content_type="application/json",
            body='{"name": "test-service", "status": "Healthy"}'
        ))

        # Navigate to Services page
        # Assuming port 9002 (as mentioned in README)
        try:
            page.goto("http://localhost:9002/services")
        except Exception as e:
            print(f"Failed to load page: {e}")
            return

        # Wait for table to load
        try:
            page.wait_for_selector("table", timeout=10000)
        except Exception as e:
            print(f"Table not found: {e}")
            page.screenshot(path="verification/error.png")
            return

        # Find Test Connection button (Activity icon)
        # It's inside a Tooltip trigger, inside the Actions cell.
        # Button has aria-label="Test Connection"
        test_btn = page.locator('button[aria-label="Test Connection"]')

        # Take screenshot before click
        os.makedirs("verification", exist_ok=True)
        page.screenshot(path="verification/before_click.png")

        # Click it
        test_btn.click()

        # Check for toast? Toast might be hard to capture in screenshot if it animates.
        # But we can wait for text "Service Healthy".
        try:
            page.wait_for_selector("text=Service Healthy", timeout=5000)
        except Exception as e:
             print(f"Toast not found: {e}")

        # Take screenshot after click
        page.screenshot(path="verification/after_click.png")
        print("Verification complete")

        browser.close()

if __name__ == "__main__":
    verify_test_connection_button()
