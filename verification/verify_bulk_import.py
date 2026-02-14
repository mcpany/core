from playwright.sync_api import sync_playwright
import time
import json

def run(playwright):
    browser = playwright.chromium.launch(headless=True)
    context = browser.new_context()
    page = context.new_page()

    # 1. Start
    page.goto("http://localhost:9002/upstream-services")

    # 2. Click Bulk Import
    page.get_by_role("button", name="Bulk Import").click()

    # Wait for Dialog
    page.wait_for_selector("h2:has-text('Bulk Service Import')")

    # Wait for wizard content
    page.wait_for_selector("text=Paste JSON")

    time.sleep(1) # Animation wait

    page.screenshot(path="verification/step1_input.png")

    # 3. Fill JSON
    service_config = [
        {
            "name": "screenshot-service",
            "httpService": { "address": "http://example.com" }
        }
    ]
    page.get_by_label("Service Configuration (JSON)").fill(json.dumps(service_config))

    # 4. Review
    page.get_by_role("button", name="Review").click()

    # 5. Wait for Validation
    # Wait for table to appear
    page.wait_for_selector("table")
    # Wait for Valid status
    page.wait_for_selector("text=Valid")

    time.sleep(0.5)
    page.screenshot(path="verification/step2_validation.png")

    # 6. Import
    page.get_by_role("button", name="Import").click()

    # 7. Result
    page.wait_for_selector("text=Import Complete")

    time.sleep(0.5)
    page.screenshot(path="verification/step4_result.png")

    browser.close()

with sync_playwright() as playwright:
    run(playwright)
