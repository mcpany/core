from playwright.sync_api import sync_playwright
import time

def run():
    with sync_playwright() as p:
        print("Launching browser...")
        browser = p.chromium.launch()
        page = browser.new_page()

        # 1. Navigate to Marketplace
        print("Navigating to Marketplace...")
        try:
            page.goto("http://localhost:9002/marketplace", timeout=60000)
        except Exception as e:
            print(f"Navigation failed: {e}")
            return

        # Wait for page to load
        print("Waiting for page load...")
        try:
            page.wait_for_selector("text=Marketplace", timeout=10000)
            print("Page loaded.")
        except Exception as e:
            print(f"Page load failed: {e}")
            page.screenshot(path="verification/page_load_failed.png")
            return

        # Screenshot before clicking
        page.screenshot(path="verification/before_click.png")

        # 2. Click Create Config
        print("Clicking Create Config...")
        try:
            create_btn = page.locator("button", has_text="Create Config")
            if create_btn.is_visible():
                print("Button is visible.")
                create_btn.click()
            else:
                print("Button is NOT visible.")
        except Exception as e:
            print(f"Failed to click Create Config: {e}")
            page.screenshot(path="verification/create_button_failed.png")
            return

        # Wait for wizard to load
        print("Waiting for wizard...")
        try:
            # Check for Dialog title
            page.wait_for_selector("text=Create Upstream Service Config", timeout=10000)
            print("Wizard opened.")
        except Exception as e:
            print(f"Wizard did not open: {e}")
            page.screenshot(path="verification/wizard_failed.png")
            return

        # 3. Select a template
        print("Selecting PostgreSQL template...")
        try:
            page.locator("button#service-template").click()
            page.locator("div[role='option']", has_text="PostgreSQL").click()
        except Exception as e:
             print(f"Failed to select template: {e}")
             page.screenshot(path="verification/template_selection_failed.png")
             return

        # 4. Click Next
        print("Clicking Next...")
        try:
            page.locator("button", has_text="Next").click()
        except Exception as e:
            print(f"Failed to click Next: {e}")
            return

        # 5. Verify Step 2
        print("Verifying form...")
        try:
            page.wait_for_selector("text=Configuration", timeout=5000)

            if page.is_visible("text=Connection URL"):
                print("SUCCESS: Found 'Connection URL' field from schema.")
            else:
                print("FAILURE: 'Connection URL' field not found.")
        except Exception as e:
            print(f"Failed to find Configuration header: {e}")

        # Take screenshot
        print("Taking screenshot...")
        page.screenshot(path="verification/wizard_schema_form.png")

        browser.close()

if __name__ == "__main__":
    run()
