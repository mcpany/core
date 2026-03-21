import subprocess
import os

try:
    env = os.environ.copy()
    env["PATH"] = "/app/build/env/bin:" + env["PATH"]
    result = subprocess.run(["make", "lint"], cwd="/app/server", env=env, capture_output=True, text=True)
    print("STDOUT", result.stdout)
    print("STDERR", result.stderr)
except Exception as e:
    print(e)
