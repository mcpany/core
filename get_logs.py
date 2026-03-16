import urllib.request
import json
import zipfile
import io

url = "https://api.github.com/repos/mcpany/core/actions/runs/23142831993/logs"
req = urllib.request.Request(url)
# Will 401 without auth, but we can try
try:
    with urllib.request.urlopen(req) as response:
        with open("logs.zip", "wb") as f:
            f.write(response.read())

        with zipfile.ZipFile("logs.zip", 'r') as zip_ref:
            for name in zip_ref.namelist():
                if "bazel-test" in name:
                    with zip_ref.open(name) as f:
                        lines = f.read().decode('utf-8').split('\n')
                        for line in lines:
                            if "FAIL:" in line or "FAILED:" in line or "Error:" in line or "FAIL\t" in line:
                                print(line.strip())
except Exception as e:
    print(f"Error: {e}")
