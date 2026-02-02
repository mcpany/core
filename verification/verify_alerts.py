from playwright.sync_api import sync_playwright, expect

def run():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # 1. Navigate to Alerts page
        print("Navigating to Alerts page...")
        page.goto("http://localhost:9002/alerts")

        # Check if we need login (url might change to /login)
        if "/login" in page.url:
            print("Redirected to login. Logging in...")
            # We need credentials. The server log output the password.
            # But the server allows localhost without auth if no API key is set?
            # If UI redirects, it means UI thinks it needs auth?
            # Or backend returned 401.
            # Let's try to proceed. If login needed, I'll need to parse the password from logs or reset it.
            # Or I can set MCPANY_API_KEY and use it?
            # But UI needs to know how to send it.
            pass

        # 2. Verify page title
        expect(page.get_by_role("heading", name="Alerts & Incidents")).to_be_visible(timeout=10000)
        print("Alerts page loaded.")

        # 3. Create Rule
        print("Opening Create Rule dialog...")
        page.get_by_role("button", name="New Alert Rule").click()

        print("Filling form...")
        page.get_by_label("Name").fill("Test Rule Playwright")
        page.get_by_label("Condition").fill("cpu > 90")

        # Service selector might be tricky with Shadcn/Radix UI.
        # It's usually a button trigger.
        # We can leave default "All Services" or try to select.
        # page.get_by_text("Select service").click()
        # page.get_by_role("option", name="weather-service").click()

        print("Submitting...")
        page.get_by_role("button", name="Create Rule", exact=True).click()

        # 4. Verify Toast
        print("Verifying success toast...")
        expect(page.get_by_text("Rule Created")).to_be_visible()
        expect(page.get_by_text("Alert rule has been successfully created.")).to_be_visible()

        # 5. Take Screenshot
        print("Taking screenshot...")
        page.screenshot(path="verification/alerts_verification.png")

        browser.close()

if __name__ == "__main__":
    run()
