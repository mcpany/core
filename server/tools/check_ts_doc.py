# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

"""
This script checks for missing documentation (JSDoc) on exported symbols in TypeScript files.
It enforces the presence of 'Side Effects:' section in the docstring.
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

    # Regex for exported functions/classes/interfaces/types/consts
    pattern = re.compile(r'^\s*export\s+(default\s+)?(function|class|interface|type|const|enum)\s+([a-zA-Z0-9_]+)')

    missing_docs = []

    for i, line in enumerate(lines):
        match = pattern.match(line)
        if match:
            name = match.group(3)
            # Check for docstring above
            has_doc = False
            doc_lines = []
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
                    # Found end of docstring
                    # Walk backwards to capture lines until /**
                    k = j
                    while k >= 0:
                        l = lines[k].strip()
                        doc_lines.insert(0, l) # Prepend
                        if l.startswith('/**'):
                            has_doc = True
                            break
                        k -= 1
                    break
                break

            if not has_doc:
                missing_docs.append((i + 1, name, "Missing docstring"))
            else:
                doc_text = "\n".join(doc_lines)
                validate_ts_doc(doc_text, name, i + 1, missing_docs)

    return missing_docs

def validate_ts_doc(doc_text, name, line_num, missing_docs):
    """
    Validates the structure of the docstring.
    """
    # Check for Summary (at least some text that is not a tag)
    # This is a weak check, just ensuring it's not empty or just tags.

    # Check for Side Effects section
    if "Side Effects:" not in doc_text:
        missing_docs.append((line_num, name, "Missing 'Side Effects:' section"))

def main():
    """
    Main function to walk the directory and check all applicable files.
    Exits with status code 1 if any missing documentation is found.
    """
    root_dir = 'ui/src'
    has_errors = False

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
                        print(f"{filepath}:{line_num}: {msg} for exported symbol {name}")

    if has_errors:
        sys.exit(1)

if __name__ == "__main__":
    main()
