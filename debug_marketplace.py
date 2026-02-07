from playwright.sync_api import sync_playwright

def debug_marketplace():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            page.goto("http://localhost:9002/marketplace")
            page.wait_for_load_state("networkidle")
            page.screenshot(path="debug_marketplace_1.png")

            page.click("text=Public")
            page.wait_for_timeout(1000) # Wait for animation/render
            page.screenshot(path="debug_marketplace_2.png")

        except Exception as e:
            print(e)
        finally:
            browser.close()

if __name__ == "__main__":
    debug_marketplace()
