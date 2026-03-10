import os
import time
from playwright.sync_api import Page, expect, sync_playwright

def test_logs_regex_search(page: Page):
  """
  Verifies that the logs page regex search button appears.
  """
  print("Navigating to logs page...")
  page.goto("http://localhost:9002/logs")

  print("Waiting for page to load...")
  page.wait_for_load_state("networkidle")
  page.wait_for_timeout(2000)

  # Type something in search
  search_input = page.get_by_placeholder("Search logs...")
  search_input.wait_for(state="visible", timeout=10000)
  search_input.fill("test[A-Z]+")

  # Click the regex toggle button
  regex_btn = page.get_by_title("Use Regular Expression")
  regex_btn.wait_for(state="visible", timeout=10000)
  regex_btn.click()

  print("Taking screenshot of logs page...")
  os.makedirs("/home/jules/verification", exist_ok=True)
  page.screenshot(path="/home/jules/verification/logs_regex.png", full_page=True)

if __name__ == "__main__":
  with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    context = browser.new_context(viewport={"width": 1280, "height": 720})
    page = context.new_page()
    try:
      test_logs_regex_search(page)
    except Exception as e:
      print(f"Error during verification: {e}")
      page.screenshot(path="/home/jules/verification/error.png")
      raise e
    finally:
      browser.close()