
from playwright.sync_api import Page, expect, sync_playwright

def verify_services(page: Page):
    print("Navigating to Services...")
    page.goto("http://localhost:9002/services")

    print("Verifying Services elements...")
    expect(page.get_by_text("Payment Gateway")).to_be_visible()
    # Expect at least one enabled status
    expect(page.get_by_text("Enabled").first).to_be_visible()

    print("Taking screenshot...")
    page.screenshot(path="verification/services_verified.png", full_page=True)

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_services(page)
        finally:
            browser.close()
