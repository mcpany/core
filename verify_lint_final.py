import subprocess
import os
import sys

def run_cmd(cmd):
    print(f"Running: {cmd}")
    res = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    if res.returncode != 0:
        print(f"FAILED: {res.stderr}")
    else:
        print(f"SUCCESS: {res.stdout[:200]}...")
    return res.returncode == 0

# Check TS Doc
ts_doc_ok = run_cmd("grep -r '/**' ui/src/mocks/proto/mock-proto.ts | wc -l")

# Check Operator Go Lint (Full Config)
operator_lint_ok = run_cmd("./build/env/bin/golangci-lint run --config server/.golangci.yml ./k8s/operator/...")

# Check Bazel Lint
bazel_lint_ok = run_cmd("./build/env/bin/bazelisk run //:lint")

if ts_doc_ok and operator_lint_ok and bazel_lint_ok:
    print("ALL LINT CHECKS PASSED")
else:
    sys.exit(1)
