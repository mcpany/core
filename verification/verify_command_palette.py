from playwright.sync_api import Page, expect, sync_playwright
import time

def verify_command_palette(page: Page):
    # Navigate to the dashboard
    page.set_viewport_size({"width": 1280, "height": 720})
    page.goto("http://localhost:9002")

    # Verify Sidebar
    expect(page.locator("header")).to_be_visible()

    # Click sidebar search button
    page.locator('button:has-text("Search")').click()

    # Wait for dialog
    expect(page.get_by_role("dialog")).to_be_visible()

    # Type in search
    page.get_by_placeholder("Type a command or search...").fill("Settings")

    time.sleep(1) # Wait for animation

    # Take screenshot
    page.screenshot(path="verification/command_palette_verified.png")
    print("Screenshot saved to verification/command_palette_verified.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_command_palette(page)
        finally:
            browser.close()
