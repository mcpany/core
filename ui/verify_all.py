from playwright.sync_api import sync_playwright
import time

def verify_ui():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(viewport={"width": 1920, "height": 1080})
        page = context.new_page()

        # 1. Dashboard
        page.goto("http://localhost:3000")
        page.wait_for_selector("text=Dashboard")
        page.screenshot(path=".audit/ui/dashboard.png")
        print("Captured dashboard.png")

        # 2. Services
        page.goto("http://localhost:3000/services")
        page.wait_for_selector("text=Services")
        # Click toggle (simulate interaction)
        toggle = page.locator("button[role='switch']").first
        if toggle.is_visible():
            toggle.click()
            time.sleep(0.5) # Wait for animation/update
        page.screenshot(path=".audit/ui/services.png")
        print("Captured services.png")

        # 3. Tools
        page.goto("http://localhost:3000/tools")
        page.wait_for_selector("text=Tools")
        page.screenshot(path=".audit/ui/tools.png")
        print("Captured tools.png")

        # 4. Resources
        page.goto("http://localhost:3000/resources")
        page.wait_for_selector("text=Resources")
        page.screenshot(path=".audit/ui/resources.png")
        print("Captured resources.png")

        # 5. Prompts
        page.goto("http://localhost:3000/prompts")
        page.wait_for_selector("text=Prompts")
        page.screenshot(path=".audit/ui/prompts.png")
        print("Captured prompts.png")

        # 6. Profiles
        page.goto("http://localhost:3000/settings/profiles")
        page.wait_for_selector("text=Profiles")
        page.screenshot(path=".audit/ui/profiles.png")
        print("Captured profiles.png")

        # 7. Webhooks
        page.goto("http://localhost:3000/settings/webhooks")
        page.wait_for_selector("text=Webhooks")
        page.screenshot(path=".audit/ui/webhooks.png")
        print("Captured webhooks.png")

        # 8. Middleware
        page.goto("http://localhost:3000/settings/middleware")
        page.wait_for_selector("text=Middleware")
        page.screenshot(path=".audit/ui/middleware.png")
        print("Captured middleware.png")

        browser.close()

if __name__ == "__main__":
    verify_ui()
