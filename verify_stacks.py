from playwright.sync_api import sync_playwright
import time

def verify_stacks():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        # Use existing context to share state if needed, or new context
        context = browser.new_context()
        page = context.new_page()

        # Navigate to Stacks page
        # Assuming UI runs on 3000 (default) or whatever I set.
        # I will set PORT=3000 for UI.
        page.goto("http://localhost:3000/stacks")

        # Wait for loading to finish
        try:
            page.wait_for_selector("text=Loading stacks...", state="detached", timeout=5000)
        except:
            pass

        # Check for "New Stack" button
        page.wait_for_selector("text=New Stack")

        # Take screenshot of list
        page.screenshot(path="verification_stacks_list.png")

        # Click New Stack
        page.click("text=New Stack")

        # Wait for dialog
        page.wait_for_selector("text=Create New Stack")

        # Take screenshot of dialog
        page.screenshot(path="verification_create_stack_dialog.png")

        browser.close()

if __name__ == "__main__":
    verify_stacks()
