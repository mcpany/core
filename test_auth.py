import subprocess
import json
import time
import urllib.request
import urllib.error

def test_api():
    print("Starting backend...")
    proc = subprocess.Popen(
        ["go", "run", "./cmd/server", "start", "--json-rpc-port", "8070"],
        cwd="server",
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL
    )

    # Wait for the server to be ready
    for _ in range(10):
        try:
            req = urllib.request.Request("http://127.0.0.1:8070/healthz?api_key=test-token")
            with urllib.request.urlopen(req) as response:
                if response.status == 200:
                    print("Backend is up!")
                    break
        except urllib.error.URLError:
            pass
        time.sleep(1)
    else:
        print("Backend failed to start.")
        proc.kill()
        return

    payload = {
      "name": "e2e-auth-display-cred",
      "authentication": {
        "oauth2": {
          "client_id": { "plainText": "test-client-id-e2e" },
          "scopes": "read,write,admin",
          "token_url": "https://auth.example.com/oauth/token"
        }
      }
    }

    print(f"Sending payload: {json.dumps(payload)}")
    data = json.dumps(payload).encode('utf-8')
    req = urllib.request.Request("http://127.0.0.1:8070/api/v1/credentials", data=data, headers={"X-API-Key": "test-token", "Content-Type": "application/json"})

    try:
        with urllib.request.urlopen(req) as response:
            print(f"Status Code: {response.status}")
            print(f"Response: {response.read().decode('utf-8')}")
    except urllib.error.HTTPError as e:
        print(f"Status Code: {e.code}")
        print(f"Response: {e.read().decode('utf-8')}")
    except Exception as e:
        print(f"Error: {e}")

    # Cleanup
    proc.kill()

test_api()
