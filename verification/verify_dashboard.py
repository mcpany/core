# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

from playwright.sync_api import sync_playwright

def verify(page):
    page.goto("http://localhost:9002")
    # Wait for dashboard to load
    page.wait_for_selector("text=Dashboard")

    # Wait for Top Tools widget to appear
    # Ideally checking for "Top Tools" text
    page.wait_for_selector("text=Top Tools", timeout=10000)

    page.screenshot(path="verification/dashboard.png")
    print("Screenshot saved to verification/dashboard.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify(page)
        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="verification/error.png")
        finally:
            browser.close()
