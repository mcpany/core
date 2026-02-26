# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

"""
This script checks for missing documentation (JSDoc) on exported symbols in TypeScript files.
It enforces the Gold Standard structure: Summary, Side Effects.
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
        A list of tuples containing (line_number, symbol_name, error_message) for missing docs.
    """
    with open(filepath, 'r') as f:
        content = f.read()

    lines = content.split('\n')

    pattern = re.compile(r'^\s*export\s+(default\s+)?(function|class|interface|type|const|enum)\s+([a-zA-Z0-9_]+)')

    errors = []

    for i, line in enumerate(lines):
        match = pattern.match(line)
        if match:
            name = match.group(3)
            # Check for docstring above
            doc_lines = []
            has_doc = False
            j = i - 1
            while j >= 0:
                prev = lines[j].strip()
                if not prev:
                    j -= 1
                    continue
                if prev.startswith('@'): # decorators
                    j -= 1
                    continue
                if prev.endswith('*/'):
                    has_doc = True
                    # Read backwards until /**
                    k = j
                    while k >= 0:
                        l = lines[k].strip()
                        doc_lines.insert(0, l)
                        if l.startswith('/**'):
                            break
                        k -= 1
                break

            if not has_doc:
                errors.append((i + 1, name, "missing docstring"))
            else:
                doc_text = "\n".join(doc_lines)
                if "Summary:" not in doc_text:
                     # Check if there is at least some text description at the top
                     # But strict mode says "Summary: ..."
                     # Let's enforce "Summary:" for now to match client.ts pattern
                     # or allow implicit summary (first lines).
                     # client.ts has explicit "Summary:".
                     errors.append((i + 1, name, "missing 'Summary:' section"))

                if "Side Effects:" not in doc_text:
                    errors.append((i + 1, name, "missing 'Side Effects:' section"))

    return errors

def main():
    """
    Main function to walk the directory and check all applicable files.
    Exits with status code 1 if any missing documentation is found.
    """
    if len(sys.argv) > 1:
        root_dir = sys.argv[1]
    else:
        root_dir = 'ui/src'

    has_errors = False

    # If root_dir is a file
    if os.path.isfile(root_dir):
        missing = check_file(root_dir)
        if missing:
            has_errors = True
            for line_num, name, msg in missing:
                print(f"{root_dir}:{line_num}: symbol {name} {msg}")
        if has_errors:
            sys.exit(1)
        return

    for dirpath, dirnames, filenames in os.walk(root_dir):
        # skip node_modules
        if 'node_modules' in dirnames:
            dirnames.remove('node_modules')

        for filename in filenames:
            if filename.endswith('.ts') or filename.endswith('.tsx'):
                if filename.endswith('.d.ts') or filename.endswith('.test.ts') or filename.endswith('.test.tsx'):
                    continue

                filepath = os.path.join(dirpath, filename)
                missing = check_file(filepath)

                if missing:
                    has_errors = True
                    for line_num, name, msg in missing:
                        print(f"{filepath}:{line_num}: symbol {name} {msg}")

    if has_errors:
        sys.exit(1)

if __name__ == "__main__":
    main()
