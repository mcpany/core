from playwright.sync_api import sync_playwright, expect

def verify_ui():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(viewport={'width': 1280, 'height': 720})

        # 1. Dashboard
        print("Verifying Dashboard...")
        try:
            page = context.new_page()
            page.goto("http://localhost:9002/")
            page.wait_for_load_state("networkidle")
            # Take screenshot
            page.screenshot(path="verification/dashboard.png")
            print("Dashboard screenshot taken.")
        except Exception as e:
            print(f"Dashboard verification failed: {e}")

        # 2. Playground
        print("Verifying Playground...")
        try:
            page = context.new_page()
            page.goto("http://localhost:9002/playground")
            page.wait_for_load_state("networkidle")
            page.screenshot(path="verification/playground.png")
            print("Playground screenshot taken.")
        except Exception as e:
            print(f"Playground verification failed: {e}")

        # 3. Logs
        print("Verifying Logs...")
        try:
            page = context.new_page()
            page.goto("http://localhost:9002/logs")
            page.wait_for_load_state("networkidle")
            page.screenshot(path="verification/logs.png")
            print("Logs screenshot taken.")
        except Exception as e:
            print(f"Logs verification failed: {e}")

        # 4. Alerts
        print("Verifying Alerts...")
        try:
            page = context.new_page()
            page.goto("http://localhost:9002/alerts")
            page.wait_for_load_state("networkidle")
            page.screenshot(path="verification/alerts.png")
            print("Alerts screenshot taken.")
        except Exception as e:
            print(f"Alerts verification failed: {e}")

        # 5. Config Validator
        print("Verifying Config Validator...")
        try:
            page = context.new_page()
            page.goto("http://localhost:9002/config-validator")
            page.wait_for_load_state("networkidle")
            page.screenshot(path="verification/config_validator.png")
            print("Config Validator screenshot taken.")
        except Exception as e:
            print(f"Config Validator verification failed: {e}")

        browser.close()

if __name__ == "__main__":
    verify_ui()
