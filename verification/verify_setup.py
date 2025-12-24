from playwright.sync_api import sync_playwright

def test_home(page):
    # Port changed to 9002 based on dev script
    page.goto("http://localhost:9002")
    # Wait for something to load to ensure server is up.
    page.wait_for_load_state("networkidle")
    page.screenshot(path="verification/home.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            test_home(page)
        finally:
            browser.close()
