from playwright.sync_api import Page, expect, sync_playwright
import time
import json

def verify_diagnostics(page: Page):
    page.on("console", lambda msg: print(f"Browser console: {msg.text}"))
    page.on("pageerror", lambda exc: print(f"Browser error: {exc}"))

    services_response = {
        "services": [
            {
                "name": "my-service",
                "id": "123",
                "http_service": {
                    "address": "http://localhost:8080"
                },
                "disable": False
            }
        ]
    }

    # Mock services list
    def handle_services(route):
        print(f"Handling request to {route.request.url}")
        route.fulfill(
            status=200,
            content_type="application/json",
            body=json.dumps(services_response)
        )

    page.route("**/api/v1/services", handle_services)

    # Mock diagnose endpoint
    def handle_diagnose(route):
        print(f"Handling diagnose to {route.request.url}")
        route.fulfill(
            status=200,
            content_type="application/json",
            body=json.dumps({"ServiceName": "my-service", "Status": "OK", "Message": "Service is healthy"})
        )

    page.route("**/api/v1/services/my-service/diagnose", handle_diagnose)

    page.goto("http://localhost:9002/services")

    # Wait for the service to appear
    expect(page.get_by_text("my-service")).to_be_visible(timeout=10000)

    # Click the "Actions" dropdown (MoreHorizontal icon)
    # The button has sr-only text "Open menu"
    page.get_by_role("button", name="Open menu").click()

    # Click "Diagnose"
    # Use get_by_text since role might be ambiguous in dropdown
    page.get_by_text("Diagnose").click()

    # Wait for dialog
    expect(page.get_by_role("dialog")).to_be_visible()

    # Click "Start Diagnostics"
    page.get_by_role("button", name="Start Diagnostics").click()

    # Expect to see "Service is connected and responding." which comes from our code when Status is OK
    expect(page.get_by_text("Service is connected and responding.")).to_be_visible()

    # Take screenshot
    page.screenshot(path="verification.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_diagnostics(page)
        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="error.png")
        finally:
            browser.close()
