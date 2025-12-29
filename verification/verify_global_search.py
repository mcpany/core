# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

from playwright.sync_api import Page, expect, sync_playwright
import os

def verify_global_search(page: Page):
    # Navigate to the dashboard
    page.goto("http://localhost:9002")

    # Wait for the page to load
    page.wait_for_load_state("networkidle")

    # Check if the search button/input is visible
    search_button = page.locator('button:has-text("Search")').first
    expect(search_button).to_be_visible()

    # Take a screenshot of the dashboard with the search button
    page.screenshot(path="/home/jules/verification/dashboard_with_search.png")

    # Click the search button to open the command palette
    search_button.click()

    # Wait for the dialog to appear
    dialog = page.locator('[role="dialog"]')
    expect(dialog).to_be_visible()

    # Take a screenshot of the open command palette
    page.screenshot(path="/home/jules/verification/command_palette_open.png")

    # Type "Theme"
    input_box = page.locator('[cmdk-input]')
    input_box.fill("Theme")

    # Wait for suggestions
    page.wait_for_timeout(500)

    # Screenshot with filtered results
    page.screenshot(path="/home/jules/verification/command_palette_filtered.png")

if __name__ == "__main__":
    if not os.path.exists("/home/jules/verification"):
        os.makedirs("/home/jules/verification")

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_global_search(page)
            print("Verification script completed successfully.")
        except Exception as e:
            print(f"Verification failed: {e}")
        finally:
            browser.close()
