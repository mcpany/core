import subprocess
import json

def process():
    try:
        with open("missing_files.json", "r") as f:
            files = json.load(f)

        print("Running lint output...")
        print(files)

        if len(files) == 0:
            print("Trying to generate files for missing files json to mock it...")

    except Exception as e:
        print(e)

process()
