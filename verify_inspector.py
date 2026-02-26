# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

"""
Verification script for the Inspector UI using Playwright.

This script automates the verification of the Inspector page by navigating to it,
checking for key elements, and taking a screenshot.
"""

from playwright.sync_api import Page, expect, sync_playwright
import time

def verify_inspector(page: Page):
    """
    Verifies the functionality and appearance of the Inspector page.

    Args:
        page (Page): The Playwright Page object to interact with.

    Returns:
        None.

    Raises:
        AssertionError: If any expected element is not found or visible.

    Side Effects:
        - Navigates the browser page to the Inspector URL.
        - Takes a screenshot and saves it to disk.
    """
    # Navigate to Inspector
    page.goto("http://localhost:9002/inspector")

    # Wait for the toolbar to be visible
    # We look for the "Search traces..." input
    expect(page.get_by_placeholder("Search traces (ID, Name)...")).to_be_visible()

    # We check if the Status select is visible
    # The Select component trigger has "Status" as placeholder text usually, but initially it might say "Status" if that is the placeholder.
    # In my code: <SelectValue placeholder="Status" />
    expect(page.get_by_text("Status")).to_be_visible()

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
