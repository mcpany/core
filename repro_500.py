import requests
import json

API_URL = "http://localhost:50050/api/v1"
API_KEY = "test-token"

def repro():
    headers = {"X-API-Key": API_KEY}
    collection = {
        "name": "repro-stack",
        "services": [
            {
                "name": "weather-service",
                "mcp_service": {
                    "stdio_connection": {
                        "command": "weather",
                        "container_image": "mcpany/weather-service:latest",
                        "env": {
                            "API_KEY": { "plain_text": "secret" }
                        }
                    }
                }
            }
        ]
    }

    print("Sending POST request...")
    resp = requests.post(f"{API_URL}/collections", json=collection, headers=headers)
    print(f"Status: {resp.status_code}")
    print(f"Response: {resp.text}")

if __name__ == "__main__":
    repro()
