import urllib.request
import json
import os

url = "https://api.github.com/repos/mcpany/core/actions/runs/23142047055/jobs"
req = urllib.request.Request(url)
# Add auth if needed, but actions logs might be public? We'll see.
try:
    with urllib.request.urlopen(req) as response:
        data = json.loads(response.read())
        for job in data.get('jobs', []):
            if job['name'] == 'bazel-test' and job['conclusion'] == 'failure':
                print(f"Found failed job: {job['id']}")
                log_url = f"https://api.github.com/repos/mcpany/core/actions/jobs/{job['id']}/logs"
                print(f"To view logs, you'd fetch: {log_url}")
except Exception as e:
    print(e)
