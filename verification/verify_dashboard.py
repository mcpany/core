
from playwright.sync_api import Page, expect, sync_playwright

def verify_dashboard(page: Page):
    print("Navigating to Dashboard...")
    page.goto("http://localhost:9002")

    print("Verifying Dashboard elements...")
    expect(page.get_by_text("Total Requests")).to_be_visible()
    expect(page.get_by_text("Service Health")).to_be_visible()

    print("Taking screenshot...")
    page.screenshot(path="verification/dashboard_verified.png", full_page=True)

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_dashboard(page)
        finally:
            browser.close()
