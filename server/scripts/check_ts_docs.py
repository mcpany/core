#!/usr/bin/env python3
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import os
import re
import sys

def check_ts_docs(root_dir):
    missing_docs = []
    # Regex for exported symbols
    exported_pattern = re.compile(r'^export\s+(function|const|class|type|interface|enum)\s+([a-zA-Z0-9_]+)')
    # Handle "export default function Name" or "export default function"
    exported_default_pattern = re.compile(r'^export\s+default\s+(function|class)\s+([a-zA-Z0-9_]*)')

    for dirpath, _, filenames in os.walk(root_dir):
        for filename in filenames:
            if not filename.endswith('.ts') and not filename.endswith('.tsx'):
                continue
            if filename.endswith('.test.ts') or filename.endswith('.test.tsx'):
                continue

            filepath = os.path.join(dirpath, filename)

            with open(filepath, 'r', encoding='utf-8') as f:
                lines = f.readlines()

            for i, line in enumerate(lines):
                line = line.strip()
                match = exported_pattern.match(line)
                symbol_name = ""
                symbol_type = ""

                if match:
                    symbol_type = match.group(1)
                    symbol_name = match.group(2)
                else:
                    match_default = exported_default_pattern.match(line)
                    if match_default:
                        symbol_type = match_default.group(1)
                        symbol_name = match_default.group(2) or "default"

                if symbol_name:
                    # Check for preceding comment (/** ... */)
                    has_doc = False
                    j = i - 1
                    while j >= 0:
                        prev_line = lines[j].strip()
                        if prev_line.endswith('*/'):
                            has_doc = True
                            break
                        elif prev_line == '':
                            # Allow empty lines? TSDoc usually attaches to the declaration.
                            break
                        elif prev_line.startswith('//'):
                             # Single line comments might be docs, but we want JSDoc /** */
                             # But let's check if it's a "high quality" comment.
                             # Strict mode: Only /** ... */ counts as TSDoc.
                             pass
                        else:
                            break
                        j -= 1

                    if not has_doc:
                        missing_docs.append(f"{filepath}:{i+1} {symbol_type} {symbol_name}")

    return missing_docs

if __name__ == '__main__':
    root = sys.argv[1] if len(sys.argv) > 1 else 'ui/src'
    print(f"Scanning {root} for missing TS docs...")
    missing = check_ts_docs(root)
    if missing:
        print("Found missing docs:")
        for m in missing:
            print(m)
        # Don't exit 1 yet, just report
    else:
        print("No missing docs found!")
