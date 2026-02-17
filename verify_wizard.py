from playwright.sync_api import sync_playwright
import time

def verify_wizard():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            print("Navigating to marketplace...")
            page.goto("http://localhost:9002/marketplace?wizard=open")

            # Wait for wizard to appear
            print("Waiting for wizard...")
            page.wait_for_selector("text=Create Upstream Service Config")

            # Select "PostgreSQL" from template dropdown
            # We need to find the Select trigger.
            # StepServiceType renders a Select with trigger id="service-template"
            print("Selecting template...")
            page.click("button#service-template")
            page.click("text=PostgreSQL")

            # Verify name updates (PostgreSQL template sets name)
            # Actually template sets name if config.name is empty.

            # Click Next
            print("Clicking Next...")
            page.click("button:has-text('Next')")

            # Verify SchemaForm appears
            # StepParameters should render "Service Configuration" if schema is present
            print("Verifying form...")
            page.wait_for_selector("text=Service Configuration")
            page.wait_for_selector("text=Connection URL")

            # Take screenshot
            print("Taking screenshot...")
            page.screenshot(path="verification.png")

        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="error.png")
        finally:
            browser.close()

if __name__ == "__main__":
    verify_wizard()
