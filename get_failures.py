import os, json
try:
    with open("needs.json", "r") as f:
        needs = json.load(f)
    failed_jobs = [k for k, v in needs.items() if isinstance(v, dict) and v.get("result") in ["failure", "cancelled"]]
    print(",".join(failed_jobs))
except Exception as e:
    print(f"Error parsing JSON: {e}")
