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
            page.get_by_role("tab", name="Public").click()

            print("Clicking MCP Market...")
            # Wait for the card to be visible
            # It might take a moment to switch tabs
            page.wait_for_selector("text=MCP Market", timeout=5000)
            page.click("text=MCP Market")

            print("Waiting for Linear server...")
            page.wait_for_load_state("networkidle")

            # Click Install on Linear
            page.click("text=Install")

            print("Waiting for dialog...")
            expect(page.get_by_text("Instantiate Service")).to_be_visible()

            print("Verifying Alert...")
            expect(page.get_by_text("Configuration Auto-Detected")).to_be_visible()

            print("Taking screenshot...")
            page.screenshot(path="verification_manifest.png")
            print("Success!")

        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="error_fixed.png")
        finally:
            browser.close()

if __name__ == "__main__":
    verify_manifest()
