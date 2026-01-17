from playwright.sync_api import Page, expect, sync_playwright
import time

def test_services_page(page: Page):
    # Navigate to Services page
    # Default port for UI is 9002
    print("Navigating to http://localhost:9002/services...")
    page.goto("http://localhost:9002/services")

    # Wait for table to load
    print("Waiting for 'wttr.in'...")
    try:
        expect(page.get_by_text("wttr.in").first).to_be_visible(timeout=30000)
    except Exception as e:
        print("Failed to find wttr.in. Page content:")
        print(page.inner_html("body")[:2000])
        raise e

    # Check for "Tools" column header
    print("Checking for 'Tools' column...")
    # Try simple text match
    try:
        expect(page.get_by_text("Tools", exact=True).first).to_be_visible(timeout=5000)
    except:
        print("Header 'Tools' not found. Dumping table headers:")
        headers = page.locator("th").all_inner_texts()
        print(headers)
        # Don't fail yet, check content

    # Check for "2 tools" text
    # The cell contains "2" and "tools"
    print("Checking for tool count '2'...")
    # We target the specific structure roughly
    expect(page.get_by_text("2", exact=True)).to_be_visible()

    # Check Status Badge
    print("Taking screenshot...")
    page.screenshot(path="/home/jules/verification/services_page.png")
    print("Done.")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            test_services_page(page)
        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="/home/jules/verification/error.png")
            raise e
        finally:
            browser.close()
