# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

from playwright.sync_api import Page, expect, sync_playwright

def test_health_history(page: Page):
    # 1. Navigate to the service detail page
    # Note: Ensure backend and frontend are running.
    page.goto("http://localhost:9002/upstream-services/wttr.in")

    # 2. Wait for the page to load
    expect(page.get_by_role("heading", name="wttr.in")).to_be_visible(timeout=10000)

    # 3. Wait for the Stats card and the Health History title
    expect(page.get_by_text("Health History (Last 50 checks)")).to_be_visible()

    # 4. Check for timeline presence
    # The timeline might say "No health history available." if no checks ran yet.
    # Or show bars.

    # We just want to capture the state.
    page.screenshot(path="/home/jules/verification/health_history.png")

if __name__ == "__main__":
  with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    page = browser.new_page()
    try:
      test_health_history(page)
    finally:
      browser.close()
