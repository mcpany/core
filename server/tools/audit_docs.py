#!/usr/bin/env python3
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import os
import re
import sys

# Regex patterns
FUNC_PATTERN = re.compile(r'^func\s+([A-Z]\w*)\s*\(')
METHOD_PATTERN = re.compile(r'^func\s+\([^)]+\)\s+([A-Z]\w*)\s*\(')
TYPE_PATTERN = re.compile(r'^type\s+([A-Z]\w*)\s+')
VAR_CONST_PATTERN = re.compile(r'^(var|const)\s+([A-Z]\w*)\s+')

def has_summary(doc_lines):
    return any("Summary:" in line for line in doc_lines)

def has_parameters(doc_lines):
    return any("Parameters:" in line for line in doc_lines)

def has_returns(doc_lines):
    return any("Returns:" in line for line in doc_lines)

def check_file(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()
    missing = []
    for i, line in enumerate(lines):
        line = line.rstrip()
        symbol, kind, params_str, returns_str = None, None, "", ""
        m_func = FUNC_PATTERN.match(line)
        m_method = METHOD_PATTERN.match(line)
        m_type = TYPE_PATTERN.match(line)
        m_var_const = VAR_CONST_PATTERN.match(line)
        if m_func or m_method:
            m = m_func or m_method
            symbol = m.group(1)
            kind = "func" if m_func else "method"
            start = m.end(0)
            depth, pos = 1, start
            while pos < len(line):
                if line[pos] == '(': depth += 1
                elif line[pos] == ')': depth -= 1
                if depth == 0:
                    params_str = line[start:pos]
                    break
                pos += 1
            returns_str = line[pos+1:].strip().rstrip('{').strip()
        elif m_type:
            symbol, kind = m_type.group(1), "type"
        elif m_var_const:
            symbol, kind = m_var_const.group(2), "var/const"
        if symbol:
            doc_lines, j = [], i - 1
            while j >= 0:
                prev = lines[j].strip()
                if prev.startswith('//'):
                    if not prev.startswith('//go:') and not prev.startswith('// +build'):
                        doc_lines.insert(0, prev)
                elif prev == '': break
                else: break
                j -= 1
            reasons = []
            if not doc_lines: reasons.append("Missing documentation")
            else:
                if not has_summary(doc_lines): reasons.append("Missing 'Summary:'")
                if kind in ("func", "method"):
                    if params_str.strip():
                        if not has_parameters(doc_lines): reasons.append("Missing 'Parameters:'")
                    if returns_str.strip():
                        if not has_returns(doc_lines): reasons.append("Missing 'Returns:'")
            if reasons: missing.append((i + 1, symbol, ", ".join(reasons)))
    return missing

def scan_dir(root_dir):
    count = 0
    for root, dirs, files in os.walk(root_dir):
        if "vendor" in root or "node_modules" in root or "test" in root and "pkg" not in root: continue
        for file in files:
            if file.endswith('.go') and not file.endswith('_test.go'):
                path = os.path.join(root, file)
                missing = check_file(path)
                if missing:
                    print(f"File: {path}")
                    for line, symbol, reason in missing:
                        print(f"  Line {line}: {symbol} ({reason})")
                        count += 1
    return count

if __name__ == '__main__':
    target_dirs = sys.argv[1:] if len(sys.argv) > 1 else ['server/pkg', 'server/cmd']
    print(f"Scanning for documentation violations in {target_dirs}...")
    total = sum(scan_dir(d) for d in target_dirs)
    if total > 0:
        print(f"\nFound {total} violations.")
        sys.exit(1)
    else:
        print("\nAll public interfaces are documented correctly!")
        sys.exit(0)
