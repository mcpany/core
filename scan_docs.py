#!/usr/bin/env python3
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import os
import subprocess
import sys

def run_go_check():
    print("Scanning Go files...")
    # Compile and run the enhanced check_doc.go
    # We assume server/tools/check_doc.go is updated to be stricter.
    cmd = ["go", "run", "server/tools/check_doc.go", "-strict", "server/"]
    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        print(result.stdout)
        print(result.stderr) # Print stderr too
        return result.stdout
    return ""

def run_ts_check():
    print("Scanning TypeScript files...")
    # Run the enhanced check_ts_doc.py
    cmd = ["python3", "server/tools/check_ts_doc.py"]
    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        print(result.stdout)
        print(result.stderr)
        return result.stdout
    return ""

def main():
    go_output = run_go_check()
    ts_output = run_ts_check()

    if go_output or ts_output:
        print("\n=== Documentation Gaps Found ===")
        # print(go_output) # Already printed
        # print(ts_output) # Already printed
        sys.exit(1)
    else:
        print("\n=== All Documentation Checks Passed ===")

if __name__ == "__main__":
    main()
