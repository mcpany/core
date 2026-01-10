from playwright.sync_api import Page, expect, sync_playwright
import time
import datetime
import os

def verify_playground(page: Page):
    # Logs
    page.on("console", lambda msg: print(f"Browser console: {msg.text}"))
    page.on("pageerror", lambda err: print(f"Page error: {err}"))

    print("Navigating to playground...")
    try:
        page.goto("http://localhost:9002/playground", timeout=30000)
    except Exception as e:
        print(f"Navigation failed (timeout?): {e}")

    print("Waiting for chat input...")
    chat_input = page.get_by_placeholder("Enter command or select a tool...")
    expect(chat_input).to_be_visible(timeout=10000)

    # Type a command
    msg = 'calculator {"operation": "add", "a": 5, "b": 3}'
    print(f"Filling input with: {msg}")
    chat_input.fill(msg)

    # Wait for state update
    time.sleep(1)

    # Click Send
    send_btn = page.get_by_role("button", name="Send")
    expect(send_btn).to_be_visible()

    print("Clicking send...")
    send_btn.click()

    # 3. Assert: Check if message appears
    print("Waiting for message to appear...")
    expect(page.get_by_text(msg)).to_be_visible(timeout=10000)

    print("Checking layout...")
    expect(page.get_by_text("Library")).to_be_visible()

    # 4. Screenshot
    print("Taking screenshot...")
    time.sleep(2)

    # Prepare audit path
    date_str = datetime.date.today().isoformat()
    audit_dir = f".audit/ui/{date_str}"
    os.makedirs(audit_dir, exist_ok=True)
    screenshot_path = f"{audit_dir}/mcp_playground_pro.png"

    page.screenshot(path=screenshot_path)
    print(f"Screenshot saved to {screenshot_path}")

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page(viewport={"width": 1280, "height": 800})
        try:
            verify_playground(page)
            print("Verification successful!")
        except Exception as e:
            print(f"Verification failed: {e}")
            page.screenshot(path="failure_retry.png")
            raise e
        finally:
            browser.close()
