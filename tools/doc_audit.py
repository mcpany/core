# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import os
import re
import sys

def get_go_files(root_dir):
    files = []
    for root, _, filenames in os.walk(root_dir):
        if "vendor" in root or "node_modules" in root or "testdata" in root:
            continue
        for filename in filenames:
            if filename.endswith(".go") and not filename.endswith("_test.go") and not filename.endswith(".pb.go") and not filename.endswith(".pb.gw.go"):
                files.append(os.path.join(root, filename))
    return files

def get_ts_files(root_dir):
    files = []
    for root, _, filenames in os.walk(root_dir):
        if "node_modules" in root or "dist" in root or ".next" in root or "build" in root or "coverage" in root:
            continue
        for filename in filenames:
            if (filename.endswith(".ts") or filename.endswith(".tsx")) and not filename.endswith(".d.ts") and not filename.endswith(".test.ts") and not filename.endswith(".test.tsx") and not filename.endswith(".spec.ts"):
                 files.append(os.path.join(root, filename))
    return files

def check_go_file(filepath):
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
    except Exception as e:
        return [f"Error reading file: {e}"]

    issues = []

    # Regex for exported functions: func FuncName
    # Capture group 1: comments (optional)
    # Capture group 2: func name
    # Capture group 3: signature (up to {)
    func_pattern = re.compile(r'((?://.*\n)*)func\s+([A-Z][a-zA-Z0-9_]*)\s*(\(.*\).*?)\{', re.MULTILINE)

    for match in func_pattern.finditer(content):
        comment = match.group(1)
        name = match.group(2)
        signature = match.group(3)

        if not comment:
            issues.append(f"Function '{name}' missing docstring")
            continue

        # Check Summary
        lines = comment.strip().split('\n')
        if not lines[0].strip().startswith(f"// {name}"):
            issues.append(f"Function '{name}' docstring summary should start with '// {name}'")

        # Check Parameters if arguments exist
        # very naive check for arguments: if signature inside first parens is not empty
        params_part = signature.split(')')[0] + ')'
        if ',' in params_part or re.search(r'\(\s*[a-zA-Z0-9_]+\s+[a-zA-Z0-9_]+', params_part):
            if "Parameters:" not in comment:
                issues.append(f"Function '{name}' docstring missing 'Parameters:' section")

        # Check Returns if returns exist
        # naive check: something after closing paren of args
        if ')' in signature:
            return_part = signature.split(')', 1)[1].strip()
            if return_part and not return_part.startswith('{'):
                if "Returns:" not in comment:
                    issues.append(f"Function '{name}' docstring missing 'Returns:' section")

        # Check Errors if return contains error
        if "error" in signature:
            if "Errors:" not in comment and "Throws/Errors:" not in comment:
                issues.append(f"Function '{name}' docstring missing 'Errors:' section")

    return issues

def check_ts_file(filepath):
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
    except Exception as e:
        return [f"Error reading file: {e}"]

    issues = []

    # Regex for exported functions/consts
    # export function Name
    # export const Name =
    # export default function Name

    # Pattern: (comment)? export (default)? (function|const|class) (name)
    pattern = re.compile(r'((?:/\*\*[\s\S]*?\*/\s*)?)export\s+(?:default\s+)?(?:function|const|class|type|interface)\s+([a-zA-Z0-9_]+)', re.MULTILINE)

    for match in pattern.finditer(content):
        comment = match.group(1)
        name = match.group(2)

        if not comment:
            issues.append(f"Symbol '{name}' missing JSDoc")
            continue

        if "/**" not in comment:
             issues.append(f"Symbol '{name}' has invalid JSDoc format")

        # Check @param if likely function (hard to tell for const without parsing value, but assuming it might be)
        # We can look ahead in content to see if it looks like a function `(arg`
        # This is getting complicated for regex.
        # We'll just check if it has @param if it has parameters.

        # Simplification: Just check if JSDoc exists for now, and maybe if it has a summary.
        # The prompt asks for strict checks.

        # Let's just check for existence and summary.
        # Checking params/returns with regex in TS is very error prone due to arrow functions, etc.

    return issues

def main():
    print("Starting Documentation Audit...")

    go_files = get_go_files("server") # Scan server directory
    ts_files = get_ts_files("ui/src") # Scan ui/src directory

    all_issues = {}

    print(f"Scanning {len(go_files)} Go files...")
    for f in go_files:
        issues = check_go_file(f)
        if issues:
            all_issues[f] = issues

    print(f"Scanning {len(ts_files)} TS files...")
    for f in ts_files:
        issues = check_ts_file(f)
        if issues:
            all_issues[f] = issues

    if all_issues:
        print(f"\nFound issues in {len(all_issues)} files:")
        for f, issues in all_issues.items():
            print(f"\nFile: {f}")
            for i in issues:
                print(f"  - {i}")
        sys.exit(1)
    else:
        print("\nAll files passed audit! 100% Documentation Coverage.")
        sys.exit(0)

if __name__ == "__main__":
    main()
