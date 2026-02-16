# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

from playwright.sync_api import sync_playwright

def verify_notification_center():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Login
        print("Logging in...")
        page.goto("http://localhost:9002/login", wait_until="networkidle")

        # Ensure hydration
        page.wait_for_timeout(2000)

        page.fill('input[name="username"]', 'py-admin')
        page.fill('input[name="password"]', 'password')
        page.click('button[type="submit"]')

        print("Waiting for dashboard redirect...")
        page.wait_for_url("http://localhost:9002/", timeout=15000)
        print("Logged in.")

        # Check for Bell Icon
        print("Checking for Notification Bell...")
        bell_button = page.locator('button:has(span.sr-only:has-text("Notifications"))')
        bell_button.wait_for()

        # Check for Badge (optional, depends if there are active alerts)
        # We assume there are seeded alerts
        # page.locator('span.animate-pulse').wait_for()

        # Click Bell
        print("Opening Notification Center...")
        bell_button.click()

        # Wait for Popover content
        page.wait_for_selector('h4:has-text("Notifications")')

        # Verify seeded alerts exist (e.g. "High CPU Usage")
        # server/pkg/alerts/manager.go seeds "High CPU Usage"
        print("Verifying alerts...")
        page.wait_for_selector('div:has-text("High CPU Usage")')

        # Take screenshot of open notification center
        print("Taking screenshot...")
        page.screenshot(path="verification/notification_center.png")

        # Dismiss an alert (if active)
        # Find a dismiss button inside the popover
        dismiss_buttons = page.locator('button[title="Dismiss"]')
        if dismiss_buttons.count() > 0:
            print("Dismissing an alert...")
            dismiss_buttons.first.click()
            # Verify it disappears or changes style (hard to verify style in headless without screenshot diff, but we can check count or status)
            page.wait_for_timeout(500) # Wait for React update

            # Take post-dismiss screenshot
            print("Taking post-dismiss screenshot...")
            page.screenshot(path="verification/notification_center_dismissed.png")
        else:
            print("No active alerts to dismiss.")

        browser.close()
        print("Done.")

if __name__ == "__main__":
    verify_notification_center()
