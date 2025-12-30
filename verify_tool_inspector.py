# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0


from playwright.sync_api import Page, expect, sync_playwright
import time

def verify_tool_inspector(page: Page):
    # Navigate to tools page - using port 9002 as discovered
    page.goto("http://localhost:9002/tools")

    # Wait for content to load
    page.wait_for_selector("text=get_weather")

    # Click inspect on get_weather
    inspect_btn = page.get_by_role("row", name="get_weather").get_by_role("button", name="Inspect")
    inspect_btn.click()

    # Wait for sheet to open
    page.wait_for_selector("text=Schema Definition")

    # Fill form
    page.get_by_label("location").fill("San Francisco")

    # Take screenshot of the inspector
    page.screenshot(path=".audit/ui/2025-02-18/tool_inspector_sheet.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_tool_inspector(page)
            print("Verification successful")
        except Exception as e:
            print(f"Verification failed: {e}")
        finally:
            browser.close()
