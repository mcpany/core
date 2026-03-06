# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import os
import re

def fix_file(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    missing = []
    # Regex for exported symbols: Start of line, type/func/var/const, space, Uppercase letter
    exported_pattern = re.compile(r'^(func|type|var|const)\s+([A-Z]\w*)')
    method_pattern = re.compile(r'^func\s+\([^)]+\)\s+([A-Z]\w*)')

    for i, line in enumerate(lines):
        symbol = None
        match = exported_pattern.match(line)
        if match:
            symbol = match.group(2)
        else:
            match = method_pattern.match(line)
            if match:
                symbol = match.group(1)

        if symbol:
            doc_lines = []
            j = i - 1
            while j >= 0:
                prev = lines[j].strip()
                if prev.startswith('//'):
                    if not prev.startswith('//go:'):
                        doc_lines.insert(0, prev)
                elif prev == '':
                    break
                else:
                    break
                j -= 1

            if not doc_lines:
                missing.append((i, symbol, "Missing doc"))
            else:
                doc_text = "\n".join(doc_lines)
                if line.startswith('func'):
                    if "Parameters:" not in doc_text and "Returns:" not in doc_text:
                         missing.append((i, symbol, "Non-compliant func doc"))
                elif line.startswith('type') and 'struct' in line:
                    if "Fields:" not in doc_text and "Parameters:" not in doc_text:
                         missing.append((i, symbol, "Non-compliant struct doc"))
                elif line.startswith('type') and 'interface' in line:
                    if "Methods:" not in doc_text:
                         missing.append((i, symbol, "Non-compliant interface doc"))

    return missing

def scan_dir(root_dir):
    all_missing = 0
    for root, dirs, files in os.walk(root_dir):
        for file in files:
            if file.endswith('.go') and not file.endswith('_test.go') and 'vendor' not in root and 'proto' not in root:
                path = os.path.join(root, file)
                missing = fix_file(path)
                if missing:
                    print(f"File: {path}")
                    for line, symbol, reason in missing:
                        print(f"  Line {line}: {symbol} ({reason})")
                    all_missing += len(missing)
    print(f"Total missing: {all_missing}")

if __name__ == '__main__':
    scan_dir('server/pkg')
    scan_dir('server/cmd')
