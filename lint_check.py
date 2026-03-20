import json
import os

def list_missing():
    # just run the tool
    pass

files = []
if os.path.exists("missing_files.json"):
    with open("missing_files.json", "r") as f:
        files = json.load(f)

print(f"Total files updated: {len(files)}")
