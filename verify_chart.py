from playwright.sync_api import sync_playwright
import time

def verify_health_chart():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Mock the health history API
        page.route("**/api/v1/health/history", lambda route: route.fulfill(
            status=200,
            content_type="application/json",
            body="""{
                "points": [
                    {"timestamp": "2023-10-27T10:00:00Z", "uptimePercentage": 100, "healthyServices": 10, "totalServices": 10},
                    {"timestamp": "2023-10-27T10:01:00Z", "uptimePercentage": 90, "healthyServices": 9, "totalServices": 10},
                    {"timestamp": "2023-10-27T10:02:00Z", "uptimePercentage": 50, "healthyServices": 5, "totalServices": 10},
                    {"timestamp": "2023-10-27T10:03:00Z", "uptimePercentage": 100, "healthyServices": 10, "totalServices": 10}
                ]
            }"""
        ))

        # Mock other APIs to prevent errors
        page.route("**/api/v1/dashboard/traffic", lambda route: route.fulfill(json=[]))
        page.route("**/api/v1/doctor", lambda route: route.fulfill(json={"status": "ok"}))
        page.route("**/api/v1/settings", lambda route: route.fulfill(json={}))
        page.route("**/api/v1/services", lambda route: route.fulfill(json=[]))
        page.route("**/api/v1/tools", lambda route: route.fulfill(json=[]))
        page.route("**/api/v1/alerts", lambda route: route.fulfill(json=[]))
        page.route("**/api/v1/system/status", lambda route: route.fulfill(json={"uptime_seconds": 100}))
        page.route("**/api/v1/dashboard/metrics", lambda route: route.fulfill(json=[]))
        page.route("**/api/v1/dashboard/top-tools", lambda route: route.fulfill(json=[]))
        page.route("**/api/v1/dashboard/tool-failures", lambda route: route.fulfill(json=[]))
        page.route("**/api/v1/dashboard/tool-usage", lambda route: route.fulfill(json=[]))

        # Navigate to dashboard
        page.goto("http://localhost:9002")

        # Wait for chart to render
        time.sleep(10)

        page.evaluate("window.scrollTo(0, document.body.scrollHeight)")
        time.sleep(1)
        page.screenshot(path="verification_full.png", full_page=True)
        browser.close()

if __name__ == "__main__":
    verify_health_chart()
