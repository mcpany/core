
from playwright.sync_api import Page, expect, sync_playwright

def verify_ui_features(page: Page):
    # 1. Dashboard
    page.goto("http://localhost:9002/")
    expect(page.get_by_text("Dashboard")).to_be_visible()
    expect(page.get_by_text("Total Requests")).to_be_visible()
    page.screenshot(path="verification/dashboard_py.png")
    print("Dashboard verified")

    # 2. Services
    page.goto("http://localhost:9002/services")
    expect(page.get_by_text("Services")).to_be_visible()
    expect(page.get_by_text("Payment Gateway")).to_be_visible()
    page.screenshot(path="verification/services_py.png")
    print("Services verified")

    # 3. Tools
    page.goto("http://localhost:9002/tools")
    expect(page.get_by_text("Tools")).to_be_visible()
    expect(page.get_by_text("stripe_charge")).to_be_visible()
    page.screenshot(path="verification/tools_py.png")
    print("Tools verified")

    # 4. Settings
    page.goto("http://localhost:9002/settings")
    expect(page.get_by_text("Settings")).to_be_visible()
    expect(page.get_by_text("Execution Profiles")).to_be_visible()
    page.screenshot(path="verification/settings_py.png")
    print("Settings verified")

if __name__ == "__main__":
  with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    page = browser.new_page()
    try:
      verify_ui_features(page)
    finally:
      browser.close()
