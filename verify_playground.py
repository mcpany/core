from playwright.sync_api import sync_playwright

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    context = browser.new_context()
    page = context.new_page()

    # Navigate to the dashboard
    page.goto("http://localhost:3000/playground?tool=weather-tool")

    # Wait for the page to load
    page.wait_for_selector("text=Request Builder")

    # Take a screenshot
    page.screenshot(path="verification_playground.png")

    browser.close()

with sync_playwright() as playwright:
    run(playwright)
