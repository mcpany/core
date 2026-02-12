from playwright.sync_api import sync_playwright

def run():
    print("Starting Playwright...")
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Mock the tools API
        page.route("**/api/v1/tools", lambda route: route.fulfill(
            status=200,
            content_type="application/json",
            body='{"tools": [{"name": "calculator", "description": "Perform calculations", "serviceId": "math-service"}, {"name": "search", "description": "Search the web", "serviceId": "web-service"}]}'
        ))

        # Mock other APIs to prevent errors
        page.route("**/api/v1/skills/*", lambda route: route.fulfill(status=404)) # For loadSkill
        page.route("**/api/v1/credentials", lambda route: route.fulfill(json={"credentials": []}))

        try:
            print("Navigating...")
            # Use domcontentloaded as it is faster
            page.goto("http://localhost:9002/skills/create", timeout=60000, wait_until="domcontentloaded")
            print("Navigation complete.")

            # Take debug screenshot
            page.screenshot(path="/home/jules/verification/debug_initial.png")

            print("Waiting for selector...")
            # Wait for page to load - check for the label we added "Allowed Tools"
            page.wait_for_selector("text=Allowed Tools", timeout=10000)
            print("Selector found.")

            # Click the Tool Selector - The button text depends on state. Initially "Select tools..."
            print("Clicking selector...")
            page.click("button[role='combobox']")

            # Wait for popover content
            print("Waiting for option...")
            page.wait_for_selector("text=calculator", timeout=5000)

            # Select 'calculator'
            print("Selecting option...")
            page.click("text=calculator")

            # Close popover
            page.keyboard.press("Escape")

            page.wait_for_timeout(500)

            # Take screenshot
            print("Taking final screenshot...")
            page.screenshot(path="/home/jules/verification/tool_selector.png")
            print("Screenshot taken")

        except Exception as e:
            print(f"Error: {e}")
            try:
                page.screenshot(path="/home/jules/verification/error.png")
                print("Error screenshot taken")
            except:
                print("Failed to take error screenshot")
        finally:
            browser.close()

if __name__ == "__main__":
    run()
