import os
from playwright.sync_api import Page, expect, sync_playwright

def test_dashboard_verification(page: Page):
  """
  Verifies the dashboard layout and UI elements.
  """
  print("Navigating to dashboard...")
  page.goto("http://localhost:9002")

  print("Waiting for dashboard to load...")
  page.wait_for_load_state("networkidle")
  page.wait_for_timeout(2000)

  print("Taking screenshot of dashboard...")
  os.makedirs("/home/jules/verification", exist_ok=True)
  page.screenshot(path="/home/jules/verification/dashboard.png", full_page=True)

if __name__ == "__main__":
  with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    context = browser.new_context(viewport={"width": 1280, "height": 720})
    page = context.new_page()
    try:
      test_dashboard_verification(page)
    except Exception as e:
      print(f"Error during verification: {e}")
      page.screenshot(path="/home/jules/verification/error.png")
      raise e
    finally:
      browser.close()