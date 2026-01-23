from playwright.sync_api import sync_playwright, expect

def verify_services(page):
    # 1. Arrange: Go to the Services page.
    print("Navigating to Services page...")
    page.goto("http://localhost:9002/services")

    # Wait for the page to load
    page.wait_for_load_state("networkidle")

    # 2. Act: Click "Add Service" button.
    print("Clicking Add Service...")
    page.get_by_role("button", name="Add Service").click()

    # 3. Assert: Verify Category Filters are present
    print("Verifying categories...")
    expect(page.get_by_role("button", name="AI & Memory")).to_be_visible()
    expect(page.get_by_role("button", name="Web")).to_be_visible()
    expect(page.get_by_role("button", name="Productivity")).to_be_visible()

    # 4. Assert: Verify Templates are present
    print("Verifying templates...")
    expect(page.get_by_text("Memory", exact=True)).to_be_visible()
    expect(page.get_by_text("Sequential Thinking", exact=True)).to_be_visible()
    expect(page.get_by_text("Slack", exact=True)).to_be_visible()

    # 5. Act: Filter by "AI & Memory"
    print("Filtering by AI & Memory...")
    page.get_by_role("button", name="AI & Memory").click()

    # 6. Assert: Verify Filtering
    expect(page.get_by_text("Memory", exact=True)).to_be_visible()
    expect(page.get_by_text("Sequential Thinking", exact=True)).to_be_visible()
    # Slack (Productivity) should NOT be visible
    expect(page.get_by_text("Slack", exact=True)).not_to_be_visible()

    # 7. Screenshot
    print("Taking screenshot...")
    page.screenshot(path="/home/jules/verification/services_template_selector.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_services(page)
        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="/home/jules/verification/error.png")
        finally:
            browser.close()
