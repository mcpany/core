import time
import requests
from playwright.sync_api import sync_playwright

def seed_user():
    url = "http://localhost:50050/api/v1/users"
    headers = {"X-API-Key": "test-token"}
    user = {
        "id": "verify-admin",
        "authentication": {
            "basic_auth": {
                "username": "verify-admin",
                # hash for "password"
                "password_hash": "$2a$12$KPRtQETm7XKJP/L6FjYYxuCFpTK/oRs7v9U6hWx9XFnWy6UuDqK/a"
            }
        },
        "roles": ["admin"]
    }
    try:
        requests.post(url, json={"user": user}, headers=headers)
        print("Seeded user: verify-admin")
    except Exception as e:
        print(f"Failed to seed user: {e}")

def run(playwright):
    browser = playwright.chromium.launch()
    context = browser.new_context()
    page = context.new_page()

    # 1. Login
    page.goto("http://localhost:9002/login")
    page.fill('input[name="username"]', 'verify-admin')
    page.fill('input[name="password"]', 'password')
    page.click('button[type="submit"]')
    page.wait_for_url("http://localhost:9002/")

    # 2. Go to Users
    page.goto("http://localhost:9002/users")

    # 3. Open Add User Sheet
    page.click('button:has-text("Add User")')

    # 4. Fill some data to show form
    page.fill('input[name="id"]', 'new-user')

    # 5. Take screenshot
    page.screenshot(path="/home/jules/verification/verification.png")
    print("Screenshot saved to /home/jules/verification/verification.png")

    browser.close()

if __name__ == "__main__":
    seed_user()
    with sync_playwright() as playwright:
        run(playwright)
