
import os
import json
import time
import urllib.request
import urllib.error
from playwright.sync_api import sync_playwright

def seed_data():
    base_url = "http://localhost:50050/api/v1"
    headers = {
        "X-API-Key": "test-token",
        "Content-Type": "application/json"
    }

    collection = {
        "name": "Frontend-Verify-Stack",
        "description": "Stack created for frontend verification",
        "version": "1.0.0",
        "services": []
    }

    # Try delete
    try:
        req = urllib.request.Request(
            f"{base_url}/collections/Frontend-Verify-Stack",
            method="DELETE",
            headers=headers
        )
        urllib.request.urlopen(req)
    except:
        pass

    # Seed
    print("Seeding collection...")
    data = json.dumps(collection).encode('utf-8')
    req = urllib.request.Request(
        f"{base_url}/collections",
        data=data,
        method="POST",
        headers=headers
    )
    try:
        with urllib.request.urlopen(req) as response:
            print(f"Seed response: {response.status}")
    except urllib.error.HTTPError as e:
        print(f"Seed failed: {e.code} {e.read().decode()}")

def run_verification():
    seed_data()

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Wait for frontend to be ready
        print("Waiting for frontend...")
        for i in range(30):
            try:
                page.goto("http://localhost:9002/stacks", timeout=5000)
                break
            except Exception as e:
                print(f"Waiting... {e}")
                time.sleep(2)

        print("Navigating to /stacks")
        page.goto("http://localhost:9002/stacks")

        # Wait for the card to appear
        try:
            page.wait_for_selector("text=Frontend-Verify-Stack", timeout=10000)
        except:
             print("Timeout waiting for stack text. Taking screenshot anyway.")

        # Take screenshot
        os.makedirs("verification", exist_ok=True)
        screenshot_path = os.path.abspath("verification/stacks_page.png")
        page.screenshot(path=screenshot_path)
        print(f"Screenshot saved to {screenshot_path}")

        browser.close()

if __name__ == "__main__":
    run_verification()
