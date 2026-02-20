from playwright.sync_api import sync_playwright

def verify_filesystem_ui():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        page.set_default_timeout(60000)

        try:
            # Navigate to services page
            print("Navigating...")
            page.goto("http://localhost:9002/upstream-services")

            # Open Add Service
            print("Clicking Add Service...")
            page.get_by_role("button", name="Add Service").click()

            # Wait a bit for animation
            page.wait_for_timeout(2000)

            # Take debug screenshot
            page.screenshot(path="verification/debug_sheet.png")

            # Select Custom Service
            print("Clicking Custom Service...")
            page.get_by_text("Custom Service").click()

            # Go to Connection tab
            print("Going to Connection tab...")
            page.get_by_role("tab", name="Connection").click()

            # Change Service Type
            print("Changing Type...")
            page.get_by_label("Service Type").click()
            page.get_by_role("option", name="Filesystem").click()

            # Wait for Filesystem Config to appear
            page.get_by_text("Backend Storage").wait_for()
            page.get_by_text("Mount Points").wait_for()

            # Add a mount point to show the UI in action
            page.get_by_role("button", name="Add Mount Point").click()
            page.get_by_placeholder("/workspace").first.fill("/test")

            # Take screenshot
            page.screenshot(path="verification/filesystem_ui.png")
            print("Success!")

        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="verification/error.png")
        finally:
            browser.close()

if __name__ == "__main__":
    verify_filesystem_ui()
