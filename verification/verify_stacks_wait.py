from playwright.sync_api import sync_playwright

def verify_stacks():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(
            extra_http_headers={'X-API-Key': 'test-token'}
        )
        page = context.new_page()

        print("Navigating to Stacks page...")
        page.goto("http://localhost:9002/stacks")

        print("Waiting for content to load...")
        try:
            # Wait for either the empty state or the grid
            page.wait_for_selector("text=No Stacks Found", timeout=10000)
            print("Found empty state.")
        except:
            print("Timeout waiting for 'No Stacks Found', checking for stack cards...")
            # If we seeded data, we might see cards
            try:
                page.wait_for_selector(".grid", timeout=5000)
                print("Found stack grid.")
            except:
                print("Could not find empty state or grid. Taking screenshot of current state.")

        print("Taking screenshot...")
        page.screenshot(path="verification/stacks_page_loaded.png")
        print("Screenshot saved to verification/stacks_page_loaded.png")

        browser.close()

if __name__ == "__main__":
    verify_stacks()
