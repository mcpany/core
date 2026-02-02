# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

from playwright.sync_api import Page, expect, sync_playwright
import time

def verify_dashboard_persistence(page: Page):
    print("Navigating to dashboard...")
    # 1. Arrange: Go to the Dashboard
    page.goto("http://localhost:9111")

    # Wait for dashboard to load
    print("Waiting for dashboard...")
    page.wait_for_selector("text=Dashboard", timeout=30000)

    # 2. Screenshot initial state
    page.screenshot(path="verification/1_initial.png")
    print("Initial screenshot taken.")

    # 3. Act: Hide "Recent Activity"
    print("Hiding Recent Activity...")
    # Open Layout
    page.get_by_role("button", name="Layout").click()

    # Uncheck
    widget_title = "Recent Activity"
    checkbox_label = page.locator("label", has_text=widget_title)
    if checkbox_label.count() == 0:
        print("Widget checkbox not found!")
        # maybe it's already hidden?
    else:
        checkbox_label.click()

    # Close popover
    page.keyboard.press("Escape")

    # Verify hidden
    expect(page.locator(f"text={widget_title}")).to_be_hidden()

    # 4. Screenshot hidden state
    page.screenshot(path="verification/2_hidden.png")
    print("Hidden screenshot taken.")

    # 5. Wait for debounce (2s)
    print("Waiting for save...")
    time.sleep(2)

    # 6. Reload
    print("Reloading...")
    page.reload()

    # 7. Assert: Still hidden
    page.wait_for_selector("text=Dashboard", timeout=30000)
    expect(page.locator(f"text={widget_title}")).to_be_hidden()

    # 8. Screenshot final state
    page.screenshot(path="verification/3_reloaded.png")
    print("Final screenshot taken. Success!")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_dashboard_persistence(page)
        except Exception as e:
            print(f"Verification failed: {e}")
            page.screenshot(path="verification/error.png")
        finally:
            browser.close()
