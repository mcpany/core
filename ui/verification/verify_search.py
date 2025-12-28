from playwright.sync_api import sync_playwright

def verify_global_search():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Navigate to home
        page.goto("http://localhost:9002")

        # Open Command Palette
        page.click("text=Search feature...")

        # Wait for dialog
        page.wait_for_selector('input[placeholder="Type a command or search..."]')

        # Type "Dark" to filter for theme
        page.fill('input[placeholder="Type a command or search..."]', "Dark")

        # Take screenshot
        page.screenshot(path=".audit/ui/2025-02-18/global_search_open.png")

        browser.close()

if __name__ == "__main__":
    verify_global_search()
