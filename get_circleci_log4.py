import urllib.request
import json
import sys

token = 'dummy' # CI logs are public if it's an open source PR, let's just scrape html or use the GH API anonymously

req = urllib.request.Request("https://api.github.com/repos/mcpany/core/actions/runs/23142094553/jobs")

try:
    with urllib.request.urlopen(req) as response:
        data = json.loads(response.read().decode())
        for job in data.get('jobs', []):
            if job['name'] == 'lint' or job['name'] == 'ci/circleci: lint':
                print(f"Found lint job ID: {job['id']}")
                log_url = f"https://api.github.com/repos/mcpany/core/actions/jobs/{job['id']}/logs"
                print(log_url)
except Exception as e:
    print(f"Error fetching runs: {e}")
