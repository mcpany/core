from playwright.sync_api import sync_playwright

def verify_interceptor():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        page = browser.new_page()
        page.goto("http://localhost:9111/playground")

        # Wait for tools to load
        page.wait_for_selector("text=weather-service.get_weather", timeout=30000)

        # Select tool
        page.locator(".group").filter(has_text="weather-service.get_weather").click()

        # Enable Interceptor
        page.get_by_title("Interceptor Mode (Breakpoint)").click()

        # Execute
        page.get_by_role("button", name="Execute").click()

        # Wait for dialog
        page.wait_for_selector("text=Breakpoint Hit: weather-service.get_weather")

        # Take screenshot
        page.screenshot(path="verification_interceptor.png")
        browser.close()

if __name__ == "__main__":
    verify_interceptor()
