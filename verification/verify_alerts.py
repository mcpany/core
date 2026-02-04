# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

from playwright.sync_api import sync_playwright

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            # Navigate to alerts page
            page.goto("http://localhost:9002/alerts", timeout=60000)

            # Wait for content to load (or at least the header)
            page.wait_for_selector("h2:has-text('Alerts & Incidents')")

            # Take screenshot
            page.screenshot(path="verification/alerts_page.png", full_page=True)
            print("Screenshot taken")
        except Exception as e:
            print(f"Error: {e}")
        finally:
            browser.close()

if __name__ == "__main__":
    run()
