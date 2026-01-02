# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

from playwright.sync_api import Page, expect, sync_playwright
import os
import time

def verify_log_stream(page: Page):
    # Navigate to the logs page
    # Assuming the app is running on port 9002 as per package.json "dev" script
    page.goto("http://localhost:9002/logs")

    # Wait for the "Live Stream" header to confirm page load
    expect(page.get_by_text("Live Stream")).to_be_visible(timeout=30000)

    # Wait for connection status
    expect(page.get_by_text("Connected")).to_be_visible(timeout=20000)

    # Wait for at least one log to appear
    expect(page.locator("text=INFO").first).to_be_visible(timeout=30000)

    # Take a screenshot
    screenshot_path = ".audit/ui/" + time.strftime("%Y-%m-%d") + "/log_stream.png"
    # Ensure directory exists
    os.makedirs(os.path.dirname(screenshot_path), exist_ok=True)

    page.screenshot(path=screenshot_path)
    print(f"Screenshot saved to {screenshot_path}")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_log_stream(page)
        except Exception as e:
            print(f"Verification failed: {e}")
            # Take error screenshot
            page.screenshot(path="/home/jules/verification/error.png")
            print("Error screenshot saved to /home/jules/verification/error.png")
            exit(1)
        finally:
            browser.close()
