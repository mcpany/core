from playwright.sync_api import sync_playwright

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    page = browser.new_page()
    try:
        page.goto("http://localhost:9002/stats")
        print("Navigated to http://localhost:9002/stats")

        # Wait for dashboard to load (may take a moment for hydration)
        page.wait_for_selector("text=Analytics Dashboard", timeout=30000)
        print("Found 'Analytics Dashboard' title")

        # Wait for metrics (even if 0)
        page.wait_for_selector("text=Total Requests", timeout=10000)
        print("Found 'Total Requests' card")

        # Capture screenshot
        page.screenshot(path="verification/dashboard_verification.png")
        print("Screenshot saved to verification/dashboard_verification.png")

    except Exception as e:
        print(f"Error: {e}")
        page.screenshot(path="verification/error_screenshot.png")
    finally:
        browser.close()

with sync_playwright() as playwright:
    run(playwright)
