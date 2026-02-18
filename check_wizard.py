from playwright.sync_api import sync_playwright
import time

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    page = browser.new_page()
    page.goto("http://localhost:9002/marketplace?wizard=open", timeout=60000)

    # Wait for wizard
    try:
        page.wait_for_selector("text=Create Upstream Service Config", timeout=10000)
        print("Wizard visible!")
        # Check if we can find it by role
        # dialog = page.get_by_role("dialog", name="Create Upstream Service Config")
        # if dialog.is_visible():
        #     print("Dialog role visible!")
    except Exception as e:
        print(f"Wizard not found: {e}")
        page.screenshot(path="debug_wizard.png")

    browser.close()

with sync_playwright() as playwright:
    run(playwright)
