
from playwright.sync_api import sync_playwright
import time
import datetime

def verify_log_stream():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Navigate to logs page
        # Assuming dev server running on port 9002 as per package.json "dev": "next dev --turbopack -p 9002"
        try:
            page.goto("http://localhost:9002/logs")
            print("Navigated to logs page")

            # Wait for content to load
            page.wait_for_selector('[data-testid="log-rows-container"]', timeout=30000)
            print("Log container found")

            # Wait for logs to populate
            page.wait_for_timeout(3000)

            # Take screenshot
            today = datetime.date.today().strftime("%Y-%m-%d")
            screenshot_path = f".audit/ui/{today}/log_stream.png"
            page.screenshot(path=screenshot_path)
            print(f"Screenshot saved to {screenshot_path}")

        except Exception as e:
            print(f"Error: {e}")
        finally:
            browser.close()

if __name__ == "__main__":
    verify_log_stream()
