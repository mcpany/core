from playwright.sync_api import Page, expect, sync_playwright

def verify_command_palette(page: Page):
  # 1. Arrange: Go to the homepage.
  page.goto("http://localhost:9002")

  # 2. Assert: Sidebar search button is visible
  search_button = page.get_by_text("Search...")
  expect(search_button).to_be_visible()

  # 3. Act: Open command palette using click
  search_button.click()

  # 4. Assert: Dialog is open
  dialog = page.get_by_role("dialog")
  expect(dialog).to_be_visible()

  # 5. Type query
  page.get_by_placeholder("Type a command or search...").fill("Logs")

  # 6. Screenshot
  page.screenshot(path="verification/command_palette.png")

if __name__ == "__main__":
  with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    page = browser.new_page()
    try:
      verify_command_palette(page)
    finally:
      browser.close()
