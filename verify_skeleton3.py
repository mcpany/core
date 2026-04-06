from playwright.sync_api import sync_playwright
import time

def run_cuj(page):
    page.goto("http://localhost:9005/tools")
    # Intercept API to keep it "loading"
    page.route("**/api/v1/tools", lambda route: time.sleep(10))

    # Take a screenshot right away to capture the skeleton
    page.screenshot(path="/home/jules/verification/screenshots/verification.png")
    page.wait_for_timeout(1000)

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(
            record_video_dir="/home/jules/verification/videos"
        )
        page = context.new_page()
        try:
            run_cuj(page)
        finally:
            context.close()
            browser.close()