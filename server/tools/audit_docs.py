#!/usr/bin/env python3
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import os
import re
import sys

# Regex patterns
FUNC_PATTERN = re.compile(r'^func\s+([A-Z]\w*)\s*\((.*)\)\s*(.*)\{')
METHOD_PATTERN = re.compile(r'^func\s+\([^)]+\)\s+([A-Z]\w*)\s*\((.*)\)\s*(.*)\{')
TYPE_PATTERN = re.compile(r'^type\s+([A-Z]\w*)\s+')
VAR_CONST_PATTERN = re.compile(r'^(var|const)\s+([A-Z]\w*)\s+')

def has_summary(doc_lines):
    for line in doc_lines:
        if "Summary" in line:
            return True
    return False

def has_parameters(doc_lines):
    for line in doc_lines:
        if "Parameters" in line:
            return True
    return False

def has_returns(doc_lines):
    for line in doc_lines:
        if "Returns" in line:
            return True
    return False

def check_file(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    missing = []

    for i, line in enumerate(lines):
        line = line.rstrip()
        symbol = None
        kind = None
        params_str = ""
        returns_str = ""

        m_func = FUNC_PATTERN.match(line)
        m_method = METHOD_PATTERN.match(line)
        m_type = TYPE_PATTERN.match(line)
        m_var_const = VAR_CONST_PATTERN.match(line)

        if m_func:
            symbol = m_func.group(1)
            kind = "func"
            params_str = m_func.group(2)
            returns_str = m_func.group(3)
        elif m_method:
            symbol = m_method.group(1)
            kind = "method"
            params_str = m_method.group(2)
            returns_str = m_method.group(3)
        elif m_type:
            symbol = m_type.group(1)
            kind = "type"
        elif m_var_const:
            symbol = m_var_const.group(2)
            kind = "var/const"

        if symbol:
            doc_lines = []
            j = i - 1
            while j >= 0:
                prev = lines[j].strip()
                if prev.startswith('//'):
                    if not prev.startswith('//go:') and not prev.startswith('// +build'):
                        doc_lines.insert(0, prev)
                elif prev == '':
                    break
                else:
                    break
                j -= 1

            reasons = []
            if not doc_lines:
                reasons.append("Missing documentation")
            else:
                if not has_summary(doc_lines):
                    reasons.append("Missing 'Summary'")

                if kind in ("func", "method"):
                    if not has_parameters(doc_lines):
                        reasons.append("Missing 'Parameters'")
                    if not has_returns(doc_lines):
                        reasons.append("Missing 'Returns'")

            if reasons:
                missing.append((i + 1, symbol, ", ".join(reasons)))

    return missing

def scan_dir(root_dir):
    count = 0
    for root, dirs, files in os.walk(root_dir):
        if any(x in root for x in ["vendor", "node_modules", "examples", "docs", "test", "k8s"]):
            continue
        for file in files:
            if file.endswith('.go') and not any(file.endswith(x) for x in ['_test.go', '.pb.go', '.pb.gw.go', 'zz_generated.deepcopy.go']):
                path = os.path.join(root, file)
                missing = check_file(path)
                if missing:
                    print(f"File: {path}")
                    for line, symbol, reason in missing:
                        print(f"  Line {line}: {symbol} ({reason})")
                        count += 1
    return count

if __name__ == '__main__':
    if len(sys.argv) > 1:
        target_dirs = sys.argv[1:]
    else:
        target_dirs = ['server/pkg', 'server/cmd']
    total_violations = 0
    for d in target_dirs:
        total_violations += scan_dir(d)
    if total_violations > 0:
        sys.exit(1)
    else:
        sys.exit(0)
