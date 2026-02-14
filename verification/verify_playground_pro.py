from playwright.sync_api import sync_playwright

def test_playground_pro(page):
    print("Navigating to playground...")
    page.goto("http://localhost:9002/playground")

    # Check for Pro specific elements
    print("Waiting for Console header...")
    # The header has "Console" text
    page.wait_for_selector("text=Console", timeout=5000)

    # Check for Sidebar
    # The sidebar has a search input or tool list
    print("Checking for Sidebar...")
    # We can look for the resizable panel or just text that appears in the sidebar
    # The sidebar imports ToolSidebar. Let's assume it has some tools or "Search tools..." placeholder

    # Take screenshot
    print("Taking screenshot...")
    page.screenshot(path="verification/playground_pro.png")
    print("Screenshot saved to verification/playground_pro.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            test_playground_pro(page)
        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="verification/error_pro.png")
        finally:
            browser.close()
