from playwright.sync_api import Page, expect, sync_playwright
import time
import random

def verify_wizard(page: Page):
    print("Navigating...")
    page.goto("http://localhost:9111/upstream-services")

    unique_id = random.randint(1000, 9999)
    service_name = f"verify-service-{unique_id}"

    # Open Wizard
    print("Opening wizard...")
    page.get_by_role("button", name="Bulk Import").click()

    # Wait for dialog
    print("Waiting for dialog...")
    expect(page.get_by_role("dialog")).to_be_visible()
    dialog = page.get_by_role("dialog")

    # Input Step
    print("Filling input...")
    json_data = f'[{{\"name\": \"{service_name}\", \"httpService\": {{\"address\": \"http://example.com\"}}}}]'
    dialog.locator("textarea").fill(json_data)
    dialog.get_by_role("button", name="Next", exact=True).click()

    # Validate Step
    print("Waiting for validation...")
    expect(dialog.get_by_text(service_name)).to_be_visible()

    # Take screenshot of Validation Step
    page.screenshot(path="verification/step2_validate.png")

    # Select
    print("Selecting...")
    # Radix checkbox
    cb = dialog.locator('button[role="checkbox"]').first
    if cb.get_attribute("data-state") != "checked":
        cb.click(force=True)

    dialog.get_by_role("button", name="Import Selected").click()

    # Summary Step
    print("Waiting for summary...")
    # Use heading to disambiguate from toast
    expect(dialog.get_by_role("heading", name="Import Complete")).to_be_visible()

    # Take screenshot of Summary
    page.screenshot(path="verification/step4_summary.png")
    print("Done.")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_wizard(page)
        except Exception as e:
            print(f"Error: {e}")
            page.screenshot(path="verification/error.png")
        finally:
            browser.close()
