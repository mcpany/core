from playwright.sync_api import Page, expect, sync_playwright
import time

def test_wizard(page: Page):
    # 1. Arrange: Go to the marketplace.
    for i in range(10):
        try:
            page.goto("http://localhost:9002/marketplace")
            break
        except Exception as e:
            print(f"Connection failed, retrying... {i}")
            time.sleep(2)

    # 2. Act: Click "Create Config"
    create_btn = page.get_by_role("button", name="Create Config")
    expect(create_btn).to_be_visible(timeout=30000)
    create_btn.click()

    # 3. Assert Dialog Open
    expect(page.get_by_role("dialog")).to_be_visible()
    expect(page.get_by_text("Create Upstream Service Config")).to_be_visible()

    # 4. Act: Select OpenAPI Template
    page.get_by_label("Template").click()
    page.get_by_role("option", name="OpenAPI / Swagger Import").click()

    # 5. Act: Click Next (scoped to dialog)
    page.get_by_role("dialog").get_by_role("button", name="Next").click()

    # 6. Assert: Check for OpenAPI Step
    expect(page.get_by_text("Configure OpenAPI Specification")).to_be_visible()
    expect(page.get_by_label("Specification URL")).to_be_visible()

    # 7. Act: Fill URL
    page.get_by_label("Specification URL").fill("https://petstore.swagger.io/v2/swagger.json")

    # 8. Screenshot
    page.screenshot(path="verification/wizard_openapi.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            test_wizard(page)
        except Exception as e:
            print(f"Test failed: {e}")
            page.screenshot(path="verification/error.png")
            raise e
        finally:
            browser.close()
