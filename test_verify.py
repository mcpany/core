from playwright.sync_api import sync_playwright
import os

def verify_feature():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(record_video_dir="/home/jules/verification/video")
        page = context.new_page()

        try:
            # We mock the API response for the test because the backend doesn't have data
            page.route("**/api/v1/audit/logs*", lambda route: route.fulfill(
                status=200,
                content_type="application/json",
                body="""
                {
                  "entries": [
                    {
                      "timestamp": "2023-10-27T10:00:00Z",
                      "toolName": "calculate_tax",
                      "userId": "user123",
                      "profileId": "default",
                      "arguments": "{\\"amount\\": 1000, \\"rate\\": 0.05}",
                      "result": "{\\"tax\\": 50, \\"total\\": 1050}",
                      "duration": "120ms",
                      "durationMs": 120,
                      "traceId": "trace-12345"
                    },
                    {
                      "timestamp": "2023-10-27T10:05:00Z",
                      "toolName": "create_user",
                      "userId": "admin",
                      "profileId": "default",
                      "arguments": "{\\"username\\": \\"newuser\\"}",
                      "error": "User already exists",
                      "duration": "45ms",
                      "durationMs": 45,
                      "traceId": "trace-67890"
                    }
                  ]
                }
                """
            ))

            page.route("**/api/v1/doctor", lambda route: route.fulfill(status=200, content_type="application/json", body='{"status": "ok"}'))
            page.route("**/api/v1/topology", lambda route: route.fulfill(status=200, content_type="application/json", body='{"nodes": [], "edges": []}'))
            page.route("**/api/v1/users/me", lambda route: route.fulfill(status=200, content_type="application/json", body='{"id": "test"}'))

            page.goto("http://localhost:9002/audit")
            page.wait_for_timeout(2000)

            # Find the first 'View' button and click it to open the inspector
            page.get_by_role("button", name="View").first.click()
            page.wait_for_timeout(2000)

            page.screenshot(path="/home/jules/verification/verification.png")
            page.wait_for_timeout(1000)

            # Close inspector
            page.keyboard.press("Escape")
            page.wait_for_timeout(1000)

            # Open the failed one
            page.get_by_role("button", name="View").nth(1).click()
            page.wait_for_timeout(2000)

            page.screenshot(path="/home/jules/verification/verification_error.png")
            page.wait_for_timeout(1000)


        finally:
            context.close()
            browser.close()

if __name__ == "__main__":
    os.makedirs("/home/jules/verification/video", exist_ok=True)
    verify_feature()
