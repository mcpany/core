from playwright.sync_api import Page, expect, sync_playwright
import time

def verify_global_search(page: Page):
    # Navigate to home
    page.goto("http://localhost:9002")

    # Wait for hydration
    page.wait_for_timeout(2000)

    # Open Command Palette
    page.keyboard.press("Meta+k")

    # Wait for animation
    page.wait_for_timeout(500)

    # Take screenshot of open palette
    page.screenshot(path=".audit/ui/2025-02-20/global_search_open.png")

    # Search for "Services"
    page.get_by_placeholder("Type a command or search...").fill("Services")

    # Take screenshot of search results
    page.screenshot(path=".audit/ui/2025-02-20/global_search_results.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(viewport={"width": 1280, "height": 720})
        page = context.new_page()
        try:
            verify_global_search(page)
        finally:
            browser.close()
