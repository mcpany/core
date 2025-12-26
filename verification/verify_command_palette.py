
from playwright.sync_api import sync_playwright
import time
import os

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context()
        page = context.new_page()

        try:
            print("Navigating to home page...")
            page.goto("http://localhost:9002")

            # Wait for hydration
            page.wait_for_timeout(3000)

            print("Pressing Cmd+K...")
            # Simulate Cmd+K
            page.keyboard.press("Meta+k")

            # Wait for command palette to open
            page.wait_for_timeout(1000)

            print("Taking screenshot...")
            # Take screenshot of the command palette
            # We can try to take a screenshot of the whole page, the dialog should be visible

            screenshot_path = f".audit/ui/{time.strftime('%Y-%m-%d')}/global_search.png"
            os.makedirs(os.path.dirname(screenshot_path), exist_ok=True)

            page.screenshot(path=screenshot_path)
            print(f"Screenshot saved to {screenshot_path}")

        except Exception as e:
            print(f"Error: {e}")
        finally:
            browser.close()

if __name__ == "__main__":
    run()
