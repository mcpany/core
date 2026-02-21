import os
import time
from playwright.sync_api import sync_playwright

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            print("Navigating to dashboard...")
            page.goto("http://localhost:9002")

            # Wait for dashboard content
            print("Waiting for System Health widget...")
            # The widget title is "System Health" inside a CardTitle
            page.wait_for_selector("text=System Health", timeout=30000)

            # Take screenshot
            screenshot_path = os.path.join(os.getcwd(), "verification", "dashboard.png")
            page.screenshot(path=screenshot_path)
            print(f"Screenshot saved to {screenshot_path}")

        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="verification/error.png")
        finally:
            browser.close()

if __name__ == "__main__":
    run()
