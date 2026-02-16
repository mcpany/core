# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

from playwright.sync_api import sync_playwright

def verify_users_page():
    """Verifies the functionality of the Users page by performing login and navigation.

    This function launches a headless browser, logs into the application, navigates to the Users page,
    takes a screenshot of the list, opens the 'Add User' sheet, fills a form field, and takes another screenshot.

    Raises:
        PlaywrightError: If any interaction with the browser fails.
    """
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Login
        print("Logging in...")
        page.goto("http://localhost:9002/login", wait_until="networkidle")

        # Ensure hydration by waiting for an interactive element state or a specific React attribute?
        # Just waiting a bit is safe for this verification script.
        page.wait_for_timeout(2000)

        page.fill('input[name="username"]', 'py-admin')
        page.fill('input[name="password"]', 'password')
        page.click('button[type="submit"]')

        print("Waiting for dashboard redirect...")
        page.wait_for_url("http://localhost:9002/", timeout=15000)
        print("Logged in.")

        # Go to Users
        print("Navigating to Users page...")
        page.goto("http://localhost:9002/users", wait_until="networkidle")
        page.wait_for_selector('h2:has-text("Users")')

        # Take screenshot of list
        print("Taking list screenshot...")
        page.screenshot(path="verification/users_list.png")

        # Open Sheet
        print("Opening Add User sheet...")
        page.click('button:has-text("Add User")')
        page.wait_for_selector('h2:has-text("Add New User")')

        # Fill form to show validation/state
        page.fill('input[name="id"]', 'screenshot-user')

        # Take screenshot of sheet
        print("Taking sheet screenshot...")
        page.screenshot(path="verification/users_sheet.png")

        browser.close()
        print("Done.")

if __name__ == "__main__":
    verify_users_page()
