import sys
import json
import urllib.request
import urllib.error

def main():
    try:
        url = "https://api.github.com/repos/mcpany/core/actions/runs"
        req = urllib.request.Request(url)
        with urllib.request.urlopen(req) as response:
            data = json.loads(response.read().decode())

            for run in data.get("workflow_runs", []):
                if "lint" in run.get("name", "").lower() or run.get("status") == "completed":
                    print(f"Run ID: {run['id']}, Status: {run['status']}, Conclusion: {run['conclusion']}")
                    jobs_url = run["jobs_url"]
                    jobs_req = urllib.request.Request(jobs_url)
                    with urllib.request.urlopen(jobs_req) as jobs_response:
                        jobs_data = json.loads(jobs_response.read().decode())
                        for job in jobs_data.get("jobs", []):
                            if job["conclusion"] == "failure":
                                print(f"Failed Job: {job['name']}, URL: {job['html_url']}")
                    print("-" * 20)

    except urllib.error.URLError as e:
        print(f"Error fetching data: {e}")

if __name__ == "__main__":
    main()
