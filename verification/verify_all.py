from playwright.sync_api import sync_playwright

def verify_all_pages(page):
    # Port changed to 9002 based on dev script
    base_url = "http://localhost:9002"

    # 1. Dashboard
    page.goto(base_url)
    page.wait_for_load_state("networkidle")
    page.screenshot(path="verification/dashboard.png")

    # 2. Services
    page.goto(f"{base_url}/services")
    page.wait_for_load_state("networkidle")
    page.screenshot(path="verification/services.png")

    # 3. Tools
    page.goto(f"{base_url}/tools")
    page.wait_for_load_state("networkidle")
    page.screenshot(path="verification/tools.png")

    # 4. Resources
    page.goto(f"{base_url}/resources")
    page.wait_for_load_state("networkidle")
    page.screenshot(path="verification/resources.png")

    # 5. Prompts
    page.goto(f"{base_url}/prompts")
    page.wait_for_load_state("networkidle")
    page.screenshot(path="verification/prompts.png")

    # 6. Settings
    page.goto(f"{base_url}/settings")
    page.wait_for_load_state("networkidle")
    page.screenshot(path="verification/settings.png")

    # 7. Webhooks
    page.goto(f"{base_url}/settings/webhooks")
    page.wait_for_load_state("networkidle")
    page.screenshot(path="verification/webhooks.png")

    # 8. Middleware
    page.goto(f"{base_url}/settings/middleware")
    page.wait_for_load_state("networkidle")
    page.screenshot(path="verification/middleware.png")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_all_pages(page)
        except Exception as e:
            print(e)
        finally:
            browser.close()
