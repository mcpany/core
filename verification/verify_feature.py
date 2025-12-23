
from playwright.sync_api import Page, expect, sync_playwright
import time

def verify_global_search(page: Page):
    # 1. Go to homepage
    page.goto("http://localhost:3001")

    # 2. Wait for page to load
    page.wait_for_load_state("networkidle")

    # 3. Take screenshot of homepage (should see the search icon bottom right or top right)
    page.screenshot(path="/home/jules/verification/homepage_with_search.png")

    # 4. Open search dialog via keyboard shortcut (Cmd+K)
    page.keyboard.press("Meta+k")

    # 5. Wait for dialog
    page.wait_for_selector("input[placeholder='Type a command or search...']")

    # 6. Type "Dashboard"
    page.fill("input[placeholder='Type a command or search...']", "Dashboard")

    # 7. Wait for results
    # expect(page.get_by_role("option", name="Dashboard")).to_be_visible()
    page.wait_for_selector("[role='option']")

    # 8. Take screenshot of search results
    page.screenshot(path="/home/jules/verification/global_search_results.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_global_search(page)
        except Exception as e:
            print(f"Error: {e}")
        finally:
            browser.close()
