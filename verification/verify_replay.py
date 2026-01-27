from playwright.sync_api import Page, expect, sync_playwright
import time

def verify_replay(page: Page):
    # Mock API responses

    def handle_tools(route):
        print(f"Intercepted: {route.request.url}")
        route.fulfill(json={
            "tools": [
                {
                    "name": "calculator",
                    "description": "Basic calculator",
                    "serviceId": "math-service",
                    "inputSchema": {
                        "type": "object",
                        "properties": {
                            "a": {"type": "number"},
                            "b": {"type": "number"},
                            "op": {"type": "string"}
                        }
                    }
                }
            ]
        })

    def handle_services(route):
        print(f"Intercepted: {route.request.url}")
        route.fulfill(json=[
            {
                "id": "math-service",
                "name": "Math Service",
                "status": "active"
            }
        ])

    def handle_usage(route):
        print(f"Intercepted: {route.request.url}")
        route.fulfill(json=[
            {
                "name": "calculator",
                "serviceId": "math-service",
                "totalCalls": 10,
                "successRate": 90
            }
        ])

    def handle_audit(route):
        print(f"Intercepted: {route.request.url}")
        route.fulfill(json={
            "entries": [
                {
                    "timestamp": "2023-10-27T10:00:00Z",
                    "toolName": "calculator",
                    "userId": "admin",
                    "profileId": "default",
                    "arguments": "{\"a\": 5, \"b\": 10, \"op\": \"add\"}",
                    "result": "{\"result\": 15}",
                    "error": "",
                    "duration": "100ms",
                    "durationMs": 100
                },
                 {
                    "timestamp": "2023-10-27T10:05:00Z",
                    "toolName": "calculator",
                    "userId": "admin",
                    "profileId": "default",
                    "arguments": "{\"a\": 20, \"b\": 0, \"op\": \"divide\"}",
                    "result": "",
                    "error": "division by zero",
                    "duration": "50ms",
                    "durationMs": 50
                }
            ]
        })

    # 1. Tools
    page.route("**/api/v1/tools*", handle_tools)

    # 2. Services
    page.route("**/api/v1/services*", handle_services)

    # 3. Tool Usage
    page.route("**/api/v1/dashboard/tool-usage*", handle_usage)

    # 4. Audit Logs
    page.route("**/api/v1/audit/logs*", handle_audit)

    # Mock settings and user to prevent loading issues
    page.route("**/api/v1/settings", lambda route: route.fulfill(json={}))
    page.route("**/api/v1/users/me", lambda route: route.fulfill(json={"id": "admin", "roles": ["admin"]}))
    page.route("**/api/v1/profiles", lambda route: route.fulfill(json=[]))

    print("Navigating to tools page...")
    # Go to tools page
    page.goto("http://localhost:9002/tools")

    print("Waiting for tool...")
    # Wait for tool to appear
    expect(page.get_by_text("calculator").first).to_be_visible()

    print("Clicking Inspect...")
    page.locator("tr").filter(has_text="calculator").get_by_text("Inspect").click()

    print("Waiting for dialog...")
    # Wait for dialog
    expect(page.get_by_role("dialog")).to_be_visible()

    print("Switching to metrics...")
    # Switch to Metrics tab
    page.get_by_role("tab", name="Performance & Analytics").click()

    # Wait for history to load
    page.wait_for_timeout(1000)

    # Verify history entries are visible
    # We display arguments in the list
    expect(page.get_by_text('"op": "divide"')).to_be_visible()

    # Take screenshot of History
    page.screenshot(path="verification/history_view.png")

    print("Clicking Replay...")
    # Click Replay on the first item (the one with error, or successful one)
    # The list is sorted reverse chronologically, so the latest (division by zero) is first.
    # Let's replay the division by zero one.
    replay_buttons = page.get_by_role("button", name="Replay Execution")
    replay_buttons.first.click()

    print("Verifying tab switch...")
    # Verify we switched to Test & Execute tab
    expect(page.get_by_role("tab", name="Test & Execute")).to_have_attribute("data-state", "active")

    print("Verifying arguments...")
    # Verify arguments are populated
    # {"a": 20, "b": 0, "op": "divide"}
    expect(page.locator("textarea#args")).to_contain_text('"a": 20')
    expect(page.locator("textarea#args")).to_contain_text('"op": "divide"')

    # Take screenshot of Replay result
    page.screenshot(path="verification/replay_success.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_replay(page)
            print("Verification successful!")
        except Exception as e:
            print(f"Verification failed: {e}")
            page.screenshot(path="verification/failure.png")
            raise
        finally:
            browser.close()
