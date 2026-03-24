import subprocess
import os
import json

def check():
    os.chdir('server')
    r = subprocess.run(["../build/env/bin/golangci-lint", "run", "./...", "--out-format", "json"], capture_output=True, text=True)
    if r.returncode != 0:
        print("Linter failed", r.returncode)
        try:
            d = json.loads(r.stdout)
            for issue in d.get("Issues", [])[:50]:
                print(issue["Pos"]["Filename"], issue["Pos"]["Line"], issue["Text"])
        except Exception as e:
            print("parse error", e)
            print(r.stdout[:500])
            print(r.stderr[:500])
    else:
        print("clean")
check()
