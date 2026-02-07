from playwright.sync_api import sync_playwright, expect

def verify_manifest():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            print("Navigating to Marketplace...")
            page.goto("http://localhost:9002/marketplace")
            page.wait_for_load_state("networkidle")

            print("Clicking Public tab...")
            page.click("text=Public")

            print("Clicking MCP Market...")
            # Wait for the card to be visible
            page.wait_for_selector("text=MCP Market")
            page.click("text=MCP Market")

            print("Waiting for Linear server...")
            page.wait_for_load_state("networkidle")

            # Click Install on Linear
            # Assuming there is a button "Install" near "Linear"
            # We can find the card containing "Linear" and click "Install"
            page.click("text=Install")

            print("Waiting for dialog...")
            expect(page.get_by_text("Instantiate Service")).to_be_visible()

            print("Verifying Alert...")
            expect(page.get_by_text("Configuration Auto-Detected")).to_be_visible()

            print("Verifying Environment Variable Suggestion...")
            # Check if LINEAR_API_KEY is suggested
            # The input value might be empty, but the placeholder might be suggested
            # Or the key input has the value "LINEAR_API_KEY"

            # EnvVarEditor inputs: KEY and VALUE
            # We look for an input with value "LINEAR_API_KEY"
            key_input = page.locator('input[value="LINEAR_API_KEY"]')
            expect(key_input).to_be_visible()

            # Check for "Suggested" badge
            # It's a tooltip trigger, so we might need to hover or just check if the element exists
            expect(page.get_by_text("Suggested")).to_be_visible()

            print("Taking screenshot...")
            page.screenshot(path="verification_manifest.png")
            print("Success!")

        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="error.png")
        finally:
            browser.close()

if __name__ == "__main__":
    verify_manifest()
