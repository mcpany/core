from playwright.sync_api import sync_playwright
import json
import os

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    context = browser.new_context(viewport={"width": 1280, "height": 720})
    page = context.new_page()

    page.on("console", lambda msg: print(f"PAGE LOG: {msg.text}"))
    page.on("pageerror", lambda exc: print(f"PAGE ERROR: {exc}"))
    page.on("requestfailed", lambda request: print(f"REQUEST FAILED: {request.url} {request.failure}"))

    # Mock the tools API
    def handle_tools(route):
        print("Handling tools request")
        route.fulfill(
            status=200,
            content_type="application/json",
            body=json.dumps({
                "tools": [
                    {"name": "toolA", "description": "Desc A", "serviceId": "Service Alpha", "disable": False},
                    {"name": "toolB", "description": "Desc B", "serviceId": "Service Beta", "disable": False},
                    {"name": "toolC", "description": "Desc C", "serviceId": "Service Alpha", "disable": False}
                ]
            })
        )

    # Mock APIs
    # Note: Routes added LAST take precedence in Playwright? No, Routes are matched in reverse order of creation.
    # So we should add specific routes LAST.

    # Catch-all for API to avoid 404s on unmocked endpoints
    page.route("**/api/v1/**", lambda route: route.fulfill(status=200, body="{}"))

    # Specific mocks
    page.route("**/api/v1/tools", handle_tools)

    print("Navigating to /tools...")
    try:
        page.goto("http://localhost:9002/tools")
    except Exception as e:
        print(f"Navigation failed: {e}")
        return

    print("Waiting for tools...")
    try:
        page.wait_for_selector("text=toolA", timeout=10000)
    except Exception as e:
        print(f"Timeout waiting for toolA: {e}")
        page.screenshot(path="/home/jules/verification/error_debug.png")
        return

    print("Clicking filter...")
    try:
        page.click("button[role='combobox']")
        page.wait_for_selector("text=Service Beta", timeout=5000)
    except Exception as e:
         print(f"Interaction failed: {e}")

    os.makedirs("/home/jules/verification", exist_ok=True)
    page.screenshot(path="/home/jules/verification/verification.png")
    print("Screenshot saved.")

    browser.close()

if __name__ == "__main__":
    with sync_playwright() as playwright:
        run(playwright)
