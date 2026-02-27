from playwright.sync_api import sync_playwright
import time

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch()
        page = browser.new_page()

        # Navigate to resources page
        # Note: Port 9002 is specified in package.json
        try:
            page.goto("http://localhost:9002/resources")

            # Wait for page to load - wait for the search input as a proxy for "loaded"
            page.wait_for_selector("input[placeholder='Search resources...']")

            # Take screenshot of whatever is there, even if 'app' not found yet
            page.screenshot(path="verification/resources_initial.png")

            # Try to click on 'app' folder in the sidebar tree if it exists
            # We use force=True to bypass potential overlap issues
            if page.get_by_text("app", exact=True).is_visible():
                page.get_by_text("app", exact=True).click(force=True)
                time.sleep(1)

                # Click on 'src' folder in the sidebar tree to expand/select
                if page.get_by_text("src", exact=True).is_visible():
                    page.get_by_text("src", exact=True).click(force=True)
                    time.sleep(1)

            # Take screenshot of the explorer with tree and main view populated
            page.screenshot(path="verification/resources_explorer.png")
        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="verification/error.png")
        finally:
            browser.close()

if __name__ == "__main__":
    run()
