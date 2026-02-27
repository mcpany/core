from playwright.sync_api import Page, expect, sync_playwright
import os
import time
import subprocess
import requests

def verify_tools_page_seed(page: Page):
    # 1. Seed Data via API
    # We use the python requests lib to call the backend seeding endpoint
    print("Seeding data...")
    seed_data = {
        "upstream_services": [
            {
                "id": "weather-service-seeded",
                "name": "Weather Service Seeded",
                "http_service": {
                    "address": "http://localhost:9999", # Mock
                    "tools": [
                        {
                            "name": "get_weather_seeded",
                            "description": "Get seeded weather",
                            "input_schema": {
                                "type": "object",
                                "properties": {"city": {"type": "string"}}
                            }
                        }
                    ]
                }
            }
        ]
    }

    # Retry seeding a few times as server might be starting up
    for i in range(5):
        try:
            res = requests.post("http://localhost:50050/api/v1/debug/seed", json=seed_data, headers={"X-API-Key": "test-key"})
            if res.status_code == 200:
                print("Seed successful")
                break
            else:
                print(f"Seed failed: {res.status_code} {res.text}")
        except Exception as e:
            print(f"Seed connection failed: {e}")
        time.sleep(2)

    # 2. Navigate to Tools Page
    print("Navigating to tools page...")
    # Inject API Key into localStorage or assume server has one configured
    # The app checks localStorage 'mcp_auth_token' or header.
    # We can inject via script.
    page.add_init_script("localStorage.setItem('mcp_auth_token', 'test-key');")

    page.goto("http://localhost:3000/tools")

    # 3. Verify Seeded Tool is Visible
    print("Verifying tool visibility...")
    # We look for the tool name "get_weather_seeded"
    expect(page.get_by_text("get_weather_seeded")).to_be_visible(timeout=10000)

    # 4. Take Screenshot
    print("Taking screenshot...")
    page.screenshot(path="verification_tools_page.png")

if __name__ == "__main__":
    # Ensure server is running (assumed to be running by user instructions or previous step)
    # But for safety, we assume port 50050 (backend) and 3000 (frontend) are up.
    # If not, this script will fail on navigation/request.

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            verify_tools_page_seed(page)
            print("Verification script completed.")
        except Exception as e:
            print(f"Verification failed: {e}")
            page.screenshot(path="verification_failed.png")
        finally:
            browser.close()
