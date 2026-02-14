from playwright.sync_api import sync_playwright
import time

def test_playground_tools(page):
    print("Navigating to playground...")
    page.goto("http://localhost:9002/playground")

    # Check for "Available Tools" button/sheet trigger
    # Wait for the button to be visible
    print("Waiting for Available Tools button...")
    tools_button = page.get_by_role("button", name="Available Tools")
    tools_button.wait_for(state="visible", timeout=10000)

    print("Clicking Available Tools button...")
    tools_button.click()

    # Wait for sheet to open
    print("Waiting for sheet content...")
    page.wait_for_selector("div[role='dialog']", timeout=5000)

    # Take screenshot
    print("Taking screenshot...")
    page.screenshot(path="verification/playground_tools.png")
    print("Screenshot saved to verification/playground_tools.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            test_playground_tools(page)
        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="verification/error.png")
        finally:
            browser.close()
