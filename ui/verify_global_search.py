from playwright.sync_api import sync_playwright

def verify_global_search():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Navigate to the app
        page.goto('http://localhost:9002')

        # Open Global Search with Cmd+K
        page.keyboard.press('Meta+k')

        # Check if it didn't open, try Control+k
        try:
             page.wait_for_selector('input[placeholder="Type a command or search..."]', timeout=2000)
        except:
             page.keyboard.press('Control+k')
             page.wait_for_selector('input[placeholder="Type a command or search..."]')

        # Take screenshot
        page.screenshot(path='/home/jules/.audits/ui/2025-12-29/global_search_audit.png')

        browser.close()

if __name__ == "__main__":
    verify_global_search()
