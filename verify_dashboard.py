from playwright.sync_api import Page, expect, sync_playwright
import os
import json

def verify_feature(page: Page):
  # Set some mock local storage to avoid OAuth redirect or login screen if there is one
  # Wait for server to be up
  import time
  for _ in range(10):
      try:
          import urllib.request
          urllib.request.urlopen("http://localhost:9002")
          break
      except:
          time.sleep(1)

  # Mock backend API by intercepting routes
  page.route('**/api/v1/doctor*', lambda route: route.fulfill(json={"status": "healthy"}))
  page.route('**/api/v1/users/me*', lambda route: route.fulfill(json={"id": "e2e-admin-core"}))
  page.route('**/api/v1/topology*', lambda route: route.fulfill(json={"nodes": [], "edges": []}))
  page.route('**/api/v1/services*', lambda route: route.fulfill(json=[]))
  page.route('**/api/v1/tools*', lambda route: route.fulfill(json=[]))

  # Mock traces API for recent activity widget
  page.route('**/api/v1/traces*', lambda route: route.fulfill(json=[
      {
          "id": "trace-123",
          "timestamp": new_date_iso(),
          "status": "success",
          "totalDuration": 42.5,
          "rootSpan": {
              "name": "weather_get",
              "attributes": {
                  "mcp.service_id": "weather-service",
                  "mcp.request_payload": json.dumps({"location": "San Francisco"}),
                  "mcp.response_payload": json.dumps({"temperature": 68, "condition": "Sunny"})
              }
          }
      },
      {
          "id": "trace-456",
          "timestamp": new_date_iso(),
          "status": "error",
          "totalDuration": 15.2,
          "rootSpan": {
              "name": "db_query",
              "attributes": {
                  "mcp.service_id": "postgres-db",
                  "mcp.request_payload": json.dumps({"query": "SELECT * FROM users"}),
                  "error.message": "Connection refused"
              }
          }
      }
  ]))

  page.goto("http://localhost:9002/")
  page.wait_for_timeout(1500)

  # Expand the first trace to see the payloads
  page.locator('text=weather_get').first.click()
  page.wait_for_timeout(1000)

  # Expand the second trace to see the payloads
  page.locator('text=db_query').first.click()
  page.wait_for_timeout(1000)

  page.screenshot(path="/app/verification.png")
  page.wait_for_timeout(1000)

def new_date_iso():
    from datetime import datetime
    return datetime.utcnow().isoformat() + "Z"

if __name__ == "__main__":
  with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    context = browser.new_context()
    page = context.new_page()
    try:
      verify_feature(page)
    finally:
      context.close()
      browser.close()
