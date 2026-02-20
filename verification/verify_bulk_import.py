from playwright.sync_api import sync_playwright

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        page = browser.new_page()
        try:
            print("Navigating to upstream-services...")
            page.goto("http://localhost:9002/upstream-services")

            print("Waiting for page load...")
            page.wait_for_load_state("networkidle")

            print("Taking debug screenshot...")
            page.screenshot(path="verification/debug_before_click.png")

            print("Clicking Bulk Import...")
            page.get_by_role("button", name="Bulk Import").click()

            print("Waiting for dialog...")
            page.get_by_role("heading", name="Bulk Service Import").wait_for(timeout=5000)

            print("Taking final screenshot...")
            page.screenshot(path="verification/bulk_import.png")
            print("Screenshot saved to verification/bulk_import.png")
        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="verification/error.png")
        finally:
            browser.close()

if __name__ == "__main__":
    run()
