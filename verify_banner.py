from playwright.sync_api import sync_playwright

def verify_banner():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Console logs
        page.on("console", lambda msg: print(f"Browser Console: {msg.text}"))
        page.on("requestfailed", lambda request: print(f"Request failed: {request.url} {request.failure}"))

        # Mock /doctor endpoint to return degraded status
        def handle_doctor(route):
            print(f"Intercepted request to {route.request.url}")
            route.fulfill(
                status=200,
                content_type="application/json",
                body='{"status": "degraded", "timestamp": "2024-01-01T00:00:00Z", "checks": {"internet": {"status": "failed", "message": "No connection"}, "config": {"status": "ok"}}}'
            )

        page.route("**/doctor", handle_doctor)

        # Navigate to the UI
        print("Navigating to UI...")
        try:
            page.goto("http://localhost:9002", timeout=60000)
            print("Page loaded.")
        except Exception as e:
            print(f"Failed to load page: {e}")
            browser.close()
            return

        # Wait for banner
        try:
            print("Waiting for banner...")
            page.wait_for_selector("text=System Status: Degraded", timeout=10000)
            print("Banner found.")
        except Exception as e:
            print(f"Banner not found: {e}")
            page.screenshot(path="verification_failed.png")
            browser.close()
            return

        # Screenshot
        page.screenshot(path="verification_banner.png")
        print("Screenshot saved to verification_banner.png")
        browser.close()

if __name__ == "__main__":
    verify_banner()
