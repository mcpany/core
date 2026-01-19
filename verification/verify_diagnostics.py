from playwright.sync_api import Page, expect, sync_playwright

def test_diagnostics_page(page: Page):
    # 1. Arrange: Go to Dashboard
    print("Navigating to dashboard...")
    page.goto("http://localhost:9002")

    # Wait for dashboard to load
    print("Waiting for Dashboard heading...")
    expect(page.get_by_role("heading", name="Dashboard")).to_be_visible()

    # 2. Act: Navigate to Diagnostics
    print("Clicking Diagnostics link...")
    # Sidebar might be collapsed on mobile view, but desktop usually open.
    # If not found, we might need to click menu button.
    # But usually sidebar is visible.
    diagnostics_link = page.get_by_role("link", name="Diagnostics")
    diagnostics_link.click()

    # 3. Assert: Check title
    print("Waiting for System Diagnostics heading...")
    expect(page.get_by_role("heading", name="System Diagnostics")).to_be_visible()

    # Wait for diagnostics to likely finish or at least start
    # "System Health" is the card title in diagnostics component
    print("Waiting for System Health card...")
    expect(page.get_by_text("System Health").first).to_be_visible()

    # Wait a bit for logs to populate
    print("Waiting for logs...")
    page.wait_for_timeout(5000)

    # 4. Screenshot
    print("Taking screenshot...")
    page.screenshot(path="verification/diagnostics_page.png", full_page=True)
    print("Screenshot saved.")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        # Use a larger viewport to ensure sidebar is visible
        page = browser.new_page(viewport={"width": 1280, "height": 720})
        try:
            test_diagnostics_page(page)
        except Exception as e:
            print(f"Test failed: {e}")
            page.screenshot(path="verification/error_screenshot.png")
            raise e
        finally:
            browser.close()
