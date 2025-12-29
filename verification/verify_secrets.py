from playwright.sync_api import sync_playwright

def verify_secrets_manager():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Navigate to settings page
        # Note: Port 9002 is specified in package.json dev script
        page.goto("http://localhost:9002/settings")

        # Click on Secrets tab
        page.get_by_text("Secrets & Keys").click()

        # Wait for "Add Secret" button to be visible
        page.wait_for_selector("text=Add Secret")

        # Click Add Secret
        page.click("text=Add Secret")

        # Fill form
        page.fill("input[placeholder='e.g. Production OpenAI Key']", "Test OpenAI Key")
        page.fill("input[placeholder='e.g. OPENAI_API_KEY']", "OPENAI_API_KEY")
        page.fill("input[placeholder='sk-...']", "sk-test-123456789")

        # Submit
        page.click("button:has-text('Save Secret')")

        # Verify secret is added
        page.wait_for_selector("text=Test OpenAI Key")

        # Take screenshot of the list
        page.screenshot(path="verification/api_key_manager.png")

        print("Verification successful, screenshot saved.")
        browser.close()

if __name__ == "__main__":
    verify_secrets_manager()
