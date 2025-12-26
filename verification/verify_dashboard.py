from playwright.sync_api import Page, expect, sync_playwright

def verify_dashboard(page: Page):
  # 1. Go to dashboard
  page.goto("http://localhost:9002")

  # 2. Assert dashboard title is visible
  expect(page.get_by_role("heading", name="Dashboard")).to_be_visible()

  # 3. Assert metrics are visible (assuming mock data loads)
  expect(page.get_by_text("Total Requests")).to_be_visible()

  # 4. Take screenshot
  page.screenshot(path="verification/dashboard_verify.png")

if __name__ == "__main__":
  with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    page = browser.new_page()
    try:
      verify_dashboard(page)
    finally:
      browser.close()
