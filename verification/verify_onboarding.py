import os
import requests
import time
from playwright.sync_api import sync_playwright

BASE_URL = os.environ.get("BACKEND_URL", "http://localhost:50050")
UI_URL = "http://localhost:9111"
HEADERS = {"X-API-Key": "test-token"}

def cleanup_services():
    try:
        print(f"Cleaning up services at {BASE_URL}...")
        res = requests.get(f"{BASE_URL}/api/v1/services", headers=HEADERS)
        if res.status_code == 200:
            data = res.json()
            services = data if isinstance(data, list) else data.get("services", [])
            print(f"Found {len(services)} services.")
            for s in services:
                name = s.get("name")
                print(f"Deleting service: {name}")
                requests.delete(f"{BASE_URL}/api/v1/services/{name}", headers=HEADERS)

            # Verify empty
            res = requests.get(f"{BASE_URL}/api/v1/services", headers=HEADERS)
            data = res.json()
            services = data if isinstance(data, list) else data.get("services", [])
            print(f"Services remaining: {len(services)}")
        else:
            print(f"Failed to list services: {res.status_code}")
    except Exception as e:
        print(f"Cleanup failed: {e}")

def verify_onboarding():
    cleanup_services()

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        page.on("console", lambda msg: print(f"Browser console: {msg.text}"))

        print(f"Navigating to {UI_URL}")
        try:
            page.goto(UI_URL, timeout=30000)
        except Exception as e:
            print(f"Failed to load page: {e}")
            return

        # Wait for hero
        try:
            print("Waiting for 'Welcome to MCP Any'...")
            page.wait_for_selector("text=Welcome to MCP Any", timeout=20000)
            print("Hero found!")
        except Exception as e:
            print(f"Hero not found: {e}")

        # Screenshot
        os.makedirs("verification", exist_ok=True)
        path = "verification/onboarding_hero.png"
        page.screenshot(path=path)
        print(f"Screenshot saved to {path}")

        browser.close()

if __name__ == "__main__":
    verify_onboarding()
