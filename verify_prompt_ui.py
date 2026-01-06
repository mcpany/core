
from playwright.sync_api import Page, expect, sync_playwright
import time
import os
import datetime

def verify_prompt_workbench(page: Page):
    # Navigate to prompts page
    # Default Next.js port is 3000
    page.goto("http://localhost:9002/prompts")

    # Wait for the list to load
    page.wait_for_selector("text=Prompt Library")

    # Click on a prompt if one exists (or we might need to verify the empty state if none exist)
    # Since we can't easily seed data without a backend, let's look for "No prompts found" OR a prompt.

    # We will try to wait for "Prompt Library" header first.
    expect(page.get_by_text("Prompt Library")).to_be_visible()

    # Capture the workbench state
    # We might not have prompts, but we can capture the layout.

    # Let's try to mock some data by injecting into the page via evaluate?
    # Or rely on the component's default state handling.
    # The component calls apiClient.listPrompts().
    # If the backend isn't running properly (500 or 404), it might show empty state.

    time.sleep(2) # Wait for network

    # Take screenshot
    date_str = datetime.datetime.now().strftime("%Y-%m-%d")
    filepath = f".audit/ui/{date_str}/prompt_engineer_workbench.png"

    # Ensure dir exists
    os.makedirs(os.path.dirname(filepath), exist_ok=True)

    page.screenshot(path=filepath, full_page=True)
    print(f"Screenshot saved to {filepath}")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_prompt_workbench(page)
        except Exception as e:
            print(f"Error: {e}")
        finally:
            browser.close()
