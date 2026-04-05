#!/bin/bash
./bazelisk run --config=local //server/cmd/server -- start &
PID_SERVER=$!

# For UI to build we have to run bazelisk to generate proto bindings
./bazelisk run //ui:lint

cd ui && npm run dev &
PID_UI=$!

# Wait for backend
for i in {1..30}; do
  if curl -s http://localhost:50050/api/v1/doctor | grep -q 'OK'; then
    echo "Backend is up!"
    break
  fi
  sleep 2
done

# Wait for frontend
for i in {1..30}; do
  if curl -s http://localhost:9002 > /dev/null; then
    echo "Frontend is up!"
    break
  fi
  sleep 2
done

# Seed database so the complex-tool is present
curl -X POST -H 'Content-Type: application/json' -H 'X-API-Key: test-token' http://localhost:50050/api/v1/debug/seed -d '{
    "upstream_services": [
        {
            "id": "svc_03",
            "name": "Math",
            "version": "v1.0",
            "http_service": {
                "address": "http://ui-http-echo-server:5678",
                "tools": [
                    { "name": "calculator", "description": "calc", "call_id": "calc_call" },
                    {
                        "name": "complex-tool",
                        "description": "A tool with a complex schema for testing the UI",
                        "call_id": "complex_call",
                        "input_schema": {
                            "type": "object",
                            "description": "Root object description",
                            "properties": {
                                "simple_string": { "type": "string", "description": "A simple string" },
                                "simple_number": { "type": "number" },
                                "nested_object": {
                                    "type": "object",
                                    "description": "A nested object",
                                    "properties": {
                                        "nested_string": { "type": "string" },
                                        "nested_enum": { "type": "string", "enum": ["OPTION_A", "OPTION_B", "OPTION_C"], "description": "An enum value" }
                                    },
                                    "required": ["nested_string"]
                                },
                                "array_of_objects": {
                                    "type": "array",
                                    "description": "An array of objects",
                                    "items": {
                                        "type": "object",
                                        "properties": {
                                            "item_name": { "type": "string" },
                                            "item_value": { "type": "integer" }
                                        }
                                    }
                                }
                            },
                            "required": ["simple_string", "nested_object"]
                        }
                    }
                ],
                "prompts": [],
                "calls": {
                    "calc_call": {
                        "method": "HTTP_METHOD_POST",
                        "endpoint_path": "/calc"
                    }
                }
            }
        }
    ]
}'

sleep 2

cat << 'IN_EOF' > /home/jules/verification/verify_inspector.py
import time
from playwright.sync_api import sync_playwright

def run_cuj(page):
    page.goto("http://localhost:9002/tools")
    page.wait_for_timeout(3000)

    page.request.post('http://localhost:50050/api/v1/discovery/trigger', headers={'X-API-Key': 'test-token'})
    page.wait_for_timeout(3000)

    page.reload()
    page.wait_for_timeout(3000)

    page.wait_for_selector('table tbody tr', state='visible', timeout=15000)
    page.wait_for_timeout(1000)

    inspect_btn = page.locator('tr:has-text("complex-tool") button:has-text("Inspect")').first
    if inspect_btn.is_visible():
        inspect_btn.click()
    else:
        page.locator('button:has-text("Inspect")').first.click()

    page.wait_for_timeout(2000)

    for _ in range(5):
       chevrons = page.locator('.lucide-chevron-right').all()
       if not chevrons:
          break
       for c in chevrons:
          if c.is_visible():
             c.click()
             page.wait_for_timeout(200)

    page.wait_for_timeout(1000)
    page.screenshot(path="/home/jules/verification/screenshots/verification.png")
    page.wait_for_timeout(1000)

if __name__ == "__main__":
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context(
            record_video_dir="/home/jules/verification/videos",
            viewport={'width': 1280, 'height': 800}
        )
        page = context.new_page()
        try:
            run_cuj(page)
        finally:
            context.close()
            browser.close()
IN_EOF

python /home/jules/verification/verify_inspector.py

kill $PID_SERVER
kill $PID_UI
