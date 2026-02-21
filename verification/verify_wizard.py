from playwright.sync_api import sync_playwright
import time

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(extra_http_headers={"X-API-Key": "test-token"})
        page = context.new_page()

        try:
            # Navigate to setup
            page.goto("http://localhost:9111/setup")
            page.wait_for_load_state("networkidle")

            # Welcome
            page.screenshot(path="verification/1_welcome.png")
            page.get_by_role("button", name="Get Started").click()

            # Template
            page.get_by_text("Choose a Starter Template").wait_for()
            page.screenshot(path="verification/2_template.png")

            # Select Weather
            page.get_by_text("Get real-time weather information via wttr.in.").click()

            # Configure
            page.get_by_role("button", name="Continue").wait_for()
            page.screenshot(path="verification/3_configure.png")
            page.get_by_role("button", name="Continue").click()

            # Success
            page.get_by_text("You're All Set!").wait_for(timeout=10000)
            page.screenshot(path="verification/4_success.png")

            print("Verification successful, screenshots saved.")

        except Exception as e:
            print(f"Verification failed: {e}")
            page.screenshot(path="verification/error.png")
        finally:
            browser.close()

if __name__ == "__main__":
    run()
