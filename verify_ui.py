import json
import requests
import time
import subprocess
import os

from playwright.sync_api import Page, expect, sync_playwright

def seed_db():
    payload = {
        "settings": {
            "mcp_listen_address": ":8080",
            "log_level": 1,
            "log_format": 1,
            "audit": { "enabled": True },
            "dlp": { "enabled": False },
            "gc_settings": { "interval": "1h" }
        },
        "secrets": [
            {
                "id": "sec-1",
                "name": "OpenAI Prod",
                "key": "OPENAI_API_KEY",
                "provider": "openai",
                "created_at": "2025-01-01T00:00:00Z",
                "value": "sk-real-data-test"
            },
            {
                "id": "sec-2",
                "name": "Anthropic Dev",
                "key": "ANTHROPIC_API_KEY",
                "provider": "anthropic",
                "created_at": "2025-01-01T00:00:00Z",
                "value": "sk-ant-test"
            }
        ]
    }

    for i in range(10):
        try:
            r = requests.post("http://localhost:50050/api/v1/debug/seed", json=payload)
            if r.status_code == 200:
                print("Seeded database successfully")
                return
        except Exception as e:
            pass
        print("Waiting for backend...")
        time.sleep(1)
    raise Exception("Failed to seed database")

def run_test():
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        try:
            page.goto("http://localhost:9002/settings")

            # Go to Secrets tab
            page.get_by_role("tab", name="Secrets & Keys").click()

            # Wait for list to load
            expect(page.get_by_text("OpenAI Prod")).to_be_visible()

            # Take screenshot before selection
            page.screenshot(path="secrets_before.png", full_page=True)

            # Select all
            page.get_by_role("checkbox", name="Select all").click()

            # Wait for bulk actions banner
            expect(page.get_by_text("2 selected")).to_be_visible()

            # Take screenshot of bulk actions
            page.screenshot(path="secrets_bulk_actions.png", full_page=True)

            print("Successfully took screenshots")
        finally:
            browser.close()

if __name__ == "__main__":
    seed_db()
    run_test()
