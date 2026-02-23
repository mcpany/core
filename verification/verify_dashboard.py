# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import time
from playwright.sync_api import sync_playwright

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context()
        page = context.new_page()

        try:
            print("Navigating to dashboard...")
            page.goto("http://localhost:9002")

            # Wait for some content to load or timeout
            try:
                page.wait_for_selector("div", timeout=60000)
            except:
                print("Timeout waiting for selector")

            print("Taking screenshot...")
            page.screenshot(path="verification/dashboard.png", full_page=True)
            print("Screenshot saved to verification/dashboard.png")

        except Exception as e:
            print(f"Error: {e}")
        finally:
            browser.close()

if __name__ == "__main__":
    run()
