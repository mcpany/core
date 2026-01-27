from playwright.sync_api import sync_playwright
import time

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context()
        page = context.new_page()

        # Mock API calls
        mock_service = {
            "name": "test-service",
            "id": "123",
            "version": "1.0.0",
            "rate_limit": {
                "is_enabled": True,
                "requests_per_second": 10,
                "burst": 5
            },
            "resilience": {
                "timeout": "5s",
                "circuit_breaker": {
                    "failure_rate_threshold": 0.5,
                    "consecutive_failures": 3,
                    "open_duration": "30s",
                    "half_open_requests": 1
                }
            }
        }

        def handle_services(route):
            route.fulfill(
                status=200,
                content_type="application/json",
                body='{"services": [' + str(mock_service).replace("'", '"').replace("True", "true") + ']}'
            )

        page.route("**/api/v1/services", handle_services)

        # Also need to mock getService for the edit sheet if it re-fetches
        # Based on service-editor.tsx, it uses the service passed from list initially,
        # but if it was real app it might re-fetch. Here it uses props.

        print("Navigating...")
        try:
            page.goto("http://localhost:9002/upstream-services", timeout=30000)
        except Exception as e:
            print(f"Navigation failed: {e}")
            time.sleep(2)
            page.goto("http://localhost:9002/upstream-services")

        print("Waiting for service...")
        page.wait_for_selector("text=test-service")

        print("Opening menu...")
        # Click the dropdown menu trigger
        # It's a button with "Open menu" sr-only text.
        page.get_by_role("button", name="Open menu").first.click()

        print("Clicking Edit...")
        # Click the Edit menu item
        page.get_by_role("menuitem", name="Edit").click()

        print("Waiting for sheet...")
        page.wait_for_selector("text=Edit Service")

        print("Clicking Advanced...")
        page.get_by_role("tab", name="Advanced").click()

        print("Waiting for Rate Limiting...")
        page.wait_for_selector("text=Rate Limiting")

        # Verify values are populated
        print("Verifying values...")
        # Requests / Sec input
        val = page.get_by_label("Requests / Sec").input_value()
        print(f"Requests / Sec: {val}")
        if val != "10":
            print("WARNING: Requests / Sec mismatch!")

        # Interact
        print("Interacting...")
        page.get_by_label("Requests / Sec").fill("20")

        time.sleep(1) # Wait for animation

        print("Taking screenshot...")
        page.screenshot(path="verification.png", full_page=False)
        print("Done.")

        browser.close()

if __name__ == "__main__":
    run()
