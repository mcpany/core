import time
import requests
import json
from playwright.sync_api import sync_playwright, expect

def verify_logs():
    # 1. Get tools
    try:
        resp = requests.get("http://localhost:50050/tools")
        resp.raise_for_status()
        tools = resp.json()
        if not tools:
            print("No tools found")
            return

        # Find the tool
        tool_name = "wttr.in.get_weather"

        print(f"Using tool: {tool_name}")

    except Exception as e:
        print(f"Failed to get tools: {e}")
        return

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # 2. Go to Logs page
        print("Navigating to Logs page...")
        page.goto("http://localhost:9002/logs")

        # Wait for connection (Look for "Live" badge)
        print("Waiting for connection...")
        expect(page.get_by_text("Live", exact=True)).to_be_visible(timeout=10000)

        # 3. Execute tool
        print(f"Executing tool {tool_name}...")
        exec_payload = {
            "name": tool_name,
            "arguments": {"location": "London"}
        }
        try:
            resp = requests.post("http://localhost:50050/execute", json=exec_payload)
            print(f"Execute response: {resp.status_code}")
            print(f"Execute body: {resp.text[:100]}...")
        except Exception as e:
            print(f"Execute failed: {e}")

        # 4. Wait for log entry
        print("Waiting for log entry...")
        # We look for the tool name in the logs
        # And specifically the success message

        # Use a locator that finds the text
        log_entry = page.get_by_text("Tool execution successful")
        expect(log_entry).to_be_visible(timeout=10000)

        # Take screenshot
        print("Taking screenshot...")
        page.screenshot(path="verification_logs.png")

        browser.close()

if __name__ == "__main__":
    verify_logs()
