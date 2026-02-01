from playwright.sync_api import sync_playwright
import time

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    context = browser.new_context()
    page = context.new_page()

    # Mock API
    def handle_service(route):
        route.fulfill(
            status=200,
            content_type="application/json",
            body='{"service": {"id": "test-id", "name": "Verification Service", "disable": false, "version": "1.0.0", "http_service": {"address": "http://example.com"}}}'
        )

    page.route("**/api/v1/services/test-id", handle_service)
    page.route("**/api/v1/credentials", lambda route: route.fulfill(json={"credentials": []}))
    page.route("**/api/v1/services/test-id/status", lambda route: route.fulfill(json={"metrics": {}}))
    # Mock sibling/list
    page.route("**/api/v1/services", lambda route: route.fulfill(json={"services": [{"id": "test-id", "name": "Verification Service"}]}))

    # Retry navigation as server might be starting
    for i in range(10):
        try:
            page.goto("http://localhost:9002/service/test-id")
            break
        except Exception:
            time.sleep(2)

    # Click Edit
    page.get_by_role("button", name="Edit Config").click()

    # Wait for Sheet
    page.wait_for_selector("text=Edit Service Configuration")

    # Screenshot
    page.screenshot(path="verification_sheet.png")

    browser.close()

with sync_playwright() as playwright:
    run(playwright)
