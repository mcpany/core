
from playwright.sync_api import Page, expect, sync_playwright
import os

def verify_command_palette(page: Page):
    # Navigate to the dashboard
    page.goto("http://localhost:9002/")

    # Wait for the page to load
    page.wait_for_selector("body")

    # Press Cmd+K (Meta+K) or Ctrl+K
    # Playwright's keyboard.press handles modifiers
    # Assuming Linux environment in this sandbox, but 'Control+K' is safe.
    # However, component handles both.
    page.keyboard.press("Control+k")

    # Wait for the command palette to appear
    # The input should be visible
    search_input = page.get_by_placeholder("Type a command or search...")
    expect(search_input).to_be_visible()

    # Take a screenshot of the open command palette
    page.screenshot(path="verification/global_search.png")
    print("Screenshot taken: verification/global_search.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_command_palette(page)
        finally:
            browser.close()
