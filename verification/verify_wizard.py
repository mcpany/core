from playwright.sync_api import sync_playwright
import time

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        # Use context with headers to ensure auth passes, matching playwright.config.ts behavior
        context = browser.new_context(extra_http_headers={"X-API-Key": "test-token"})
        page = context.new_page()
        try:
            # Also set localStorage just in case client uses it explicitly
            page.add_init_script("""
                localStorage.setItem('mcp_auth_token', 'test-token');
            """)

            print("Navigating...")
            page.goto("http://localhost:9002/upstream-services")

            # Open Wizard
            print("Opening Wizard...")
            page.get_by_role("button", name="Bulk Import").click()
            time.sleep(2)

            # Screenshot Step 1
            print("Screenshot Step 1...")
            page.screenshot(path="verification/wizard_step1.png")

            # Step 2
            print("Entering data...")
            page.get_by_label("Configuration Content").fill('[{"name": "test-service-viz", "httpService": {"address": "http://example.com"}}]')
            page.get_by_role("button", name="Next: Validate").click()
            time.sleep(5)

            # Screenshot Step 2
            print("Screenshot Step 2...")
            page.screenshot(path="verification/wizard_step2.png")

        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="verification/error.png")
        finally:
            browser.close()

if __name__ == "__main__":
    run()
