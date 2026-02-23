from playwright.sync_api import Page, expect, sync_playwright
import time

def verify_inspector(page: Page):
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
