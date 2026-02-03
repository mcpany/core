#!/usr/bin/env python3
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import os
import re
import sys

def check_go_docs(root_dir):
    missing_docs = []

    # Regex for top-level exported functions, types, constants, variables
    # matches: func Name, type Name, const Name, var Name
    top_level_pattern = re.compile(r'^(func|type|const|var)\s+([A-Z][a-zA-Z0-9_]*)')

    # Regex for exported methods
    # matches: func (r *Receiver) Name
    method_pattern = re.compile(r'^func\s+\([^)]+\)\s+([A-Z][a-zA-Z0-9_]*)')

    for dirpath, _, filenames in os.walk(root_dir):
        for filename in filenames:
            if not filename.endswith('.go'):
                continue
            if filename.endswith('_test.go'):
                continue
            if filename.endswith('.pb.go') or filename.endswith('.pb.gw.go'):
                continue

            filepath = os.path.join(dirpath, filename)

            with open(filepath, 'r', encoding='utf-8') as f:
                lines = f.readlines()

            for i, line in enumerate(lines):
                line = line.strip() # trim whitespace for regex matching?
                # Actually Go fmt ensures strict formatting usually, but safe to match strictly from start of line?
                # Top level definitions start at column 0.

                # Check top level
                match = top_level_pattern.match(lines[i]) # match from start of string
                symbol_name = ""
                symbol_type = ""

                if match:
                    symbol_type = match.group(1)
                    symbol_name = match.group(2)
                else:
                    # Check method
                    match_method = method_pattern.match(lines[i])
                    if match_method:
                        symbol_type = "method"
                        symbol_name = match_method.group(1)

                if symbol_name:
                    # Check for preceding comment
                    has_doc = False
                    j = i - 1
                    while j >= 0:
                        prev_line = lines[j].strip()
                        if prev_line.startswith('//'):
                            has_doc = True
                            if prev_line.startswith('//go:'):
                                pass
                            break
                        elif prev_line == '':
                            break
                        else:
                            break
                        j -= 1

                    if not has_doc:
                        missing_docs.append(f"{filepath}:{i+1} {symbol_type} {symbol_name}")

    return missing_docs

if __name__ == '__main__':
    root = sys.argv[1] if len(sys.argv) > 1 else 'server'
    print(f"Scanning {root} for missing Go docs...")
    missing = check_go_docs(root)
    if missing:
        print("Found missing docs:")
        for m in missing:
            print(m)
        # Don't exit 1, just report
    else:
        print("No missing docs found!")
