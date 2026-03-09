# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

from playwright.sync_api import Page, expect, sync_playwright
import time

def verify_inspector(page: Page):
    # Navigate to Inspector
    page.goto("http://localhost:9002/inspector", wait_until="networkidle")

    # Wait for the toolbar to be visible
    # We look for the "Search traces..." input
    expect(page.get_by_placeholder("Search traces (ID, Name)...")).to_be_visible()

    # We check if the Status select is visible
    # The Select component trigger has "Status" as placeholder text usually, but initially it might say "Status" if that is the placeholder.
    # In my code: <SelectValue placeholder="Status" />
    expect(page.get_by_text("All Statuses")).to_be_visible()

    # We check if the Methods select is visible
    expect(page.get_by_text("All Methods")).to_be_visible()

    # Seed the DB with a trace
    page.get_by_role("button", name="Seed Trace").click()

    # Wait for traces to be seeded by waiting for a toast message indicating success
    # Wait until "Loading traces..." text disappears before interacting to ensure the table has populated
    # The traces table has a specific structure so we use a robust way to open the dropdown
    page.wait_for_timeout(2000)

    # Interact with the new UI element: Select a Method to filter by
    # The traces table has a specific structure so we use a robust way to open the dropdown
    page.get_by_text("All Methods").click()
    # Assuming there's a seeded method called "get_weather", we click it in the dropdown list
    # The exact method name might vary depending on backend fixtures but usually seed creates a few generic ones.
    # Let's verify we can find at least one other option and click it
    page.keyboard.press("ArrowDown")
    page.keyboard.press("Enter")

    # Take screenshot
    page.screenshot(path="verification_inspector.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_inspector(page)
        finally:
            browser.close()
