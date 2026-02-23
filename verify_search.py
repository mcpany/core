from playwright.sync_api import sync_playwright

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    page = browser.new_page()
    try:
        page.goto("http://localhost:9002")
        # Try finding the search button more robustly
        # It's a button with text "Search..." or "Search or type >..."
        # Or look for the command kbd hint
        page.locator("button").filter(has_text="Search").first.click()

        # Wait for dialog content to appear
        page.wait_for_selector("div[role='dialog']", timeout=5000)

        # Take screenshot
        page.screenshot(path="verification_search.png")
        print("Screenshot taken")
    except Exception as e:
        print(f"Error: {e}")
        # Take screenshot on error to see what happened
        page.screenshot(path="error.png")
    finally:
        browser.close()

with sync_playwright() as playwright:
    run(playwright)
