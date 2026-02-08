from playwright.sync_api import sync_playwright

def verify_metrics():
    with sync_playwright() as p:
        # Seed data using APIRequestContext
        print("Seeding data...")
        api_request = p.request.new_context(
            base_url="http://localhost:50050",
            extra_http_headers={"X-API-Key": "test-token"}
        )
        seed_data = [
            {"time": "10:00", "requests": 100, "errors": 2, "latency": 50},
            {"time": "10:01", "requests": 120, "errors": 0, "latency": 45},
            {"time": "10:02", "requests": 80, "errors": 5, "latency": 60},
        ]
        resp = api_request.post("/api/v1/debug/seed_traffic", data=seed_data)
        if not resp.ok:
            print(f"Failed to seed data: {resp.status} {resp.status_text}")
            return
        print("Data seeded successfully.")

        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        print("Navigating to service page...")
        # Pointing to the UI port 9002
        page.goto("http://localhost:9002/service/weather-service")

        # Click Metrics Tab
        print("Clicking Metrics tab...")
        page.get_by_role("tab", name="Metrics").click()

        # Wait for chart
        print("Waiting for chart...")
        page.wait_for_selector(".recharts-surface")

        # Screenshot
        print("Taking screenshot...")
        page.screenshot(path="verification/verification.png", full_page=True)
        print("Screenshot saved to verification/verification.png")

        browser.close()

if __name__ == "__main__":
    verify_metrics()
