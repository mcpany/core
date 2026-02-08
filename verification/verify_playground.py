from playwright.sync_api import sync_playwright, expect

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            print("Navigating to playground...")
            page.goto("http://localhost:9002/playground", timeout=60000)

            # Wait for content
            print("Waiting for page content...")
            # Ideally wait for a known element
            page.wait_for_selector("text=Playground", timeout=60000)

            print("Checking for buttons...")
            # Check for Export button
            expect(page.get_by_role("button", name="Export")).to_be_visible()
            # Check for Import button
            expect(page.get_by_role("button", name="Import")).to_be_visible()

            print("Taking screenshot...")
            page.screenshot(path="verification/playground_verification.png")
            print("Done.")
        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="verification/error.png")
        finally:
            browser.close()

if __name__ == "__main__":
    run()
