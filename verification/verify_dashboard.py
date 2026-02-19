from playwright.sync_api import sync_playwright

def verify_dashboard():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Enable console logging
        page.on("console", lambda msg: print(f"Browser Console: {msg.text}"))
        page.on("pageerror", lambda exc: print(f"Browser Error: {exc}"))

        # Mock services to return empty list
        page.route("**/api/v1/services", lambda route: route.fulfill(json=[]))

        # Mock preferences to not return a saved layout (404)
        page.route("**/api/v1/user/preferences", lambda route: route.fulfill(status=404))

        # Mock other potential calls to avoid errors
        page.route("**/api/v1/system/status", lambda route: route.fulfill(json={
            "uptime_seconds": 100,
            "active_connections": 1,
            "bound_http_port": 8080,
            "bound_grpc_port": 50051,
            "version": "0.0.1",
            "security_warnings": []
        }))

        try:
            # Navigate to dashboard
            page.goto("http://localhost:9002")

            # Wait for the "Welcome to MCP Any" text
            page.wait_for_selector("text=Welcome to MCP Any", timeout=10000)

            # Take screenshot
            page.screenshot(path="verification/dashboard_empty_state.png", full_page=True)
            print("Screenshot saved to verification/dashboard_empty_state.png")

        except Exception as e:
            print(f"Verification failed: {e}")
            page.screenshot(path="verification/error.png")
        finally:
            browser.close()

if __name__ == "__main__":
    verify_dashboard()
