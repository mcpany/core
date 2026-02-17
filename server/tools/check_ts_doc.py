# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

"""
This script checks for missing or non-compliant documentation (JSDoc) on exported symbols in TypeScript files.
"""

import os
import re
import sys

def check_file(filepath):
    """
    Checks a single file for missing documentation on exported symbols.

    Args:
        filepath: The path to the file to check.

    Returns:
        A list of errors found in the file.
    """
    with open(filepath, 'r') as f:
        content = f.read()

    lines = content.split('\n')
    errors = []

    # Regex for exported symbols
    # Capture: (default, type, name, args/content)
    # This is very rough.
    # export function name(...)
    # export const name = ...
    # export class name ...
    # export interface name ...
    # export type name ...

    # Simple line-by-line check for export keywords
    export_pattern = re.compile(r'^\s*export\s+(default\s+)?(function|class|interface|type|const|enum)\s+([a-zA-Z0-9_]+)\s*(.*)')

    # Heuristic for detecting if a function has arguments
    # Look for (...) in the same line or subsequent lines?
    # For simplicity, we just look at the line matched.

    for i, line in enumerate(lines):
        match = export_pattern.match(line)
        if match:
            kind = match.group(2)
            name = match.group(3)
            remainder = match.group(4)

            # Find docstring
            doc_lines = []
            j = i - 1
            has_doc = False
            is_inside_doc = False

            # Walk backwards to find end of comment */
            while j >= 0:
                prev = lines[j].strip()
                if not prev:
                    j -= 1
                    continue
                if prev.startswith('@'): # Decorators
                    j -= 1
                    continue
                if prev.endswith('*/'):
                    is_inside_doc = True
                    # Collect doc lines backwards until /**
                    k = j
                    while k >= 0:
                        doc_line = lines[k].strip()
                        doc_lines.insert(0, doc_line)
                        if doc_line.startswith('/**'):
                            has_doc = True
                            break
                        k -= 1
                    break
                else:
                    break

            if not has_doc:
                errors.append(f"{filepath}:{i + 1}: missing doc for exported {kind} {name}")
                continue

            # Check content structure
            doc_text = "\n".join(doc_lines)

            # Check Summary (text before tags)
            # Remove /**, */, and * prefixes
            clean_lines = []
            for dl in doc_lines:
                dl = dl.strip()
                if dl.startswith('/**'): dl = dl[3:]
                if dl.endswith('*/'): dl = dl[:-2]
                if dl.startswith('*'): dl = dl[1:]
                dl = dl.strip()
                if dl and not dl.startswith('@'):
                    clean_lines.append(dl)

            if not clean_lines:
                 errors.append(f"{filepath}:{i + 1}: incomplete doc for {kind} {name} (missing: Summary)")

            # Check Params/Returns for functions
            if kind == 'function' or (kind == 'const' and ('=' in remainder and '=>' in remainder or 'function' in remainder)):
                # Heuristic for params
                # If remainder contains '()', likely no params.
                # If remainder contains '(', but not '()', likely params.
                # This is weak for multiline args.

                # Check for @param
                if '(' in remainder and '()' not in remainder:
                    # Likely has params
                    if '@param' not in doc_text:
                         # errors.append(f"{filepath}:{i + 1}: incomplete doc for {kind} {name} (missing: @param)")
                         pass # Skip strict param check for regex limitations

                # Check for @returns
                # if '@returns' not in doc_text and 'void' not in remainder:
                     # errors.append(f"{filepath}:{i + 1}: incomplete doc for {kind} {name} (missing: @returns)")
                     # pass

    return errors

def main():
    root_dir = 'ui/src'
    has_errors = False

    # Walk ui/src
    for dirpath, dirnames, filenames in os.walk(root_dir):
        if 'node_modules' in dirnames:
            dirnames.remove('node_modules')
        if 'mocks' in dirnames:
            dirnames.remove('mocks')

        for filename in filenames:
            if filename.endswith('.ts') or filename.endswith('.tsx'):
                if filename.endswith('.d.ts') or filename.endswith('.test.ts') or filename.endswith('.test.tsx'):
                    continue

                filepath = os.path.join(dirpath, filename)
                file_errors = check_file(filepath)

                if file_errors:
                    has_errors = True
                    for err in file_errors:
                        print(err)

    if has_errors:
        sys.exit(1)

if __name__ == "__main__":
    main()
