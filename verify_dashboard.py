from playwright.sync_api import sync_playwright

def verify_dashboard():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            page.goto("http://localhost:9002")
            page.wait_for_load_state("networkidle")
            print(page.title())
            page.screenshot(path="dashboard.png")
        except Exception as e:
            print(e)
        finally:
            browser.close()

if __name__ == "__main__":
    verify_dashboard()
