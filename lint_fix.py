import subprocess
import json

def get_lint_output():
    cmd = ["make", "lint"]
    try:
        import os
        result = subprocess.run(cmd, cwd="server", capture_output=True, text=True, check=False)
        print("Stdout:", result.stdout[:1000])
        print("Stderr:", result.stderr[:1000])
    except Exception as e:
        print(f"Error running linter: {e}")

get_lint_output()
