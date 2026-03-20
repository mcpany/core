import urllib.request
import json
import sys

# Public CircleCI API to get the latest pipeline for a branch
branch = "jules-14725119418485289249-6d6a4596"
url = f"https://circleci.com/api/v2/project/gh/mcpany/core/pipeline?branch={branch}"

req = urllib.request.Request(url, headers={"Accept": "application/json"})
try:
    with urllib.request.urlopen(req) as res:
        data = json.loads(res.read().decode())
        if not data.get('items'):
            print("No pipelines found")
            sys.exit(0)

        pipeline_id = data['items'][0]['id']
        print(f"Latest pipeline ID: {pipeline_id}")

        url_wf = f"https://circleci.com/api/v2/pipeline/{pipeline_id}/workflow"
        req_wf = urllib.request.Request(url_wf, headers={"Accept": "application/json"})
        with urllib.request.urlopen(req_wf) as res_wf:
            wf_data = json.loads(res_wf.read().decode())
            wf_id = wf_data['items'][0]['id']
            print(f"Workflow ID: {wf_id}")

            url_jobs = f"https://circleci.com/api/v2/workflow/{wf_id}/job"
            req_jobs = urllib.request.Request(url_jobs, headers={"Accept": "application/json"})
            with urllib.request.urlopen(req_jobs) as res_jobs:
                jobs_data = json.loads(res_jobs.read().decode())
                for job in jobs_data['items']:
                    if job['name'] == 'lint' and job['status'] == 'failed':
                        print(f"Lint job failed! Job number: {job['job_number']}")

                        # Try getting step logs (requires token usually but some are public)
                        url_steps = f"https://circleci.com/api/v1.1/project/gh/mcpany/core/{job['job_number']}"
                        req_steps = urllib.request.Request(url_steps, headers={"Accept": "application/json"})
                        try:
                            with urllib.request.urlopen(req_steps) as res_steps:
                                step_data = json.loads(res_steps.read().decode())
                                for step in step_data['steps']:
                                    if step['name'] == 'Run Lint':
                                        for action in step['actions']:
                                            if action.get('failed'):
                                                if 'output_url' in action:
                                                    out_req = urllib.request.Request(action['output_url'])
                                                    with urllib.request.urlopen(out_req) as out_res:
                                                        print("--- LINT OUTPUT ---")
                                                        for msg in json.loads(out_res.read().decode()):
                                                            print(msg['message'], end='')
                        except Exception as e:
                            print(f"Error getting step details: {e}")
except Exception as e:
    print(f"Error fetching CircleCI API: {e}")
