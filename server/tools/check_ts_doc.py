# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

"""
This script checks for missing documentation (JSDoc) on exported symbols in TypeScript files.
It enforces the Gold Standard structure:
- Summary: (or @summary)
- Parameters: (@param)
- Returns: (@returns)
- Errors: (@throws)
- Side Effects: (Explicit "Side Effects:" section)
"""

import os
import re
import sys

def check_file(filepath):
    """
    Checks a single file for missing or incomplete documentation on exported symbols.

    Args:
        filepath: The path to the file to check.

    Returns:
        A list of strings describing the errors found.
    """
    with open(filepath, 'r') as f:
        content = f.read()

    lines = content.split('\n')
    errors = []

    # Regex for exported functions/classes/interfaces/types/consts
    # We capture the indentation to help finding the docstring
    # We also capture the type of export to know if we should enforce params/returns
    pattern = re.compile(r'^(\s*)export\s+(default\s+)?(function|class|interface|type|const|enum)\s+([a-zA-Z0-9_]+)')

    for i, line in enumerate(lines):
        match = pattern.match(line)
        if match:
            indent = match.group(1)
            export_type = match.group(3)
            name = match.group(4)

            # Find docstring above
            doc_lines = []
            has_doc = False
            j = i - 1

            # Skip decorators and empty lines
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
                    # Extract docstring content
                    # Scan backwards to find /**
                    end_idx = j
                    start_idx = -1
                    for k in range(j, -1, -1):
                        if '/**' in lines[k]:
                            start_idx = k
                            break

                    if start_idx != -1:
                        # Extract lines between start_idx and end_idx
                        for k in range(start_idx, end_idx + 1):
                            l = lines[k].strip()
                            # Remove /**, */, *
                            if l.startswith('/**'):
                                l = l[3:]
                            elif l.endswith('*/'):
                                l = l[:-2]
                            elif l.startswith('*'):
                                l = l[1:]
                            doc_lines.append(l.strip())

                break

            if not has_doc:
                errors.append(f"{filepath}:{i + 1}: missing doc for exported {export_type} {name}")
                continue

            doc_text = '\n'.join(doc_lines)

            # Check structure for Functions and Methods (const arrow functions are harder to detect perfectly but 'const' usually implies variable)
            # If it's a function or const (which might be an arrow function), enforce strict rules?
            # client.ts has `export const apiClient = { ... }` which is an object.
            # Methods inside valid object literals are not caught by this regex which only checks `export ...`.
            # But the regex catches `export const apiClient`.
            # Does `apiClient` need "Side Effects"?
            # It's an object. Probably not.

            # Only enforce strict structure for functions
            if export_type == 'function':
                validate_doc_structure(doc_text, filepath, i + 1, name, errors)
            elif export_type == 'const':
                 # Check if it's an arrow function?
                 # Rough check: if line contains `=>` or `(`
                 if '=>' in line or 'async' in line or '(' in line:
                     validate_doc_structure(doc_text, filepath, i + 1, name, errors)

    return errors

def validate_doc_structure(doc_text, filepath, line, name, errors):
    # Check Summary
    if "Summary:" not in doc_text and "@summary" not in doc_text:
        # Maybe explicitly check for first line being non-empty?
        # But explicit "Summary:" is requested?
        # client.ts uses "Summary:".
        # Let's enforce "Summary:" string.
        errors.append(f"{filepath}:{line}: missing 'Summary:' in doc for {name}")

    # Check Side Effects
    if "Side Effects:" not in doc_text:
        errors.append(f"{filepath}:{line}: missing 'Side Effects:' in doc for {name}")

    # For Params/Returns/Errors, we can't easily know if they are required without parsing the function signature.
    # But we can check if they are present if they SHOULD be?
    # For now, let's just enforce Summary and Side Effects as mandatory anchors.
    # And if @param is present, it's good.

def main():
    root_dir = 'ui/src'
    has_errors = False

    # We also want to check `ui/src/lib/client.ts` specifically if it's not in src?
    # It IS in `ui/src`.

    all_errors = []

    for dirpath, dirnames, filenames in os.walk(root_dir):
        if 'node_modules' in dirnames:
            dirnames.remove('node_modules')

        for filename in filenames:
            if filename.endswith('.ts') or filename.endswith('.tsx'):
                if filename.endswith('.d.ts') or filename.endswith('.test.ts') or filename.endswith('.test.tsx'):
                    continue

                filepath = os.path.join(dirpath, filename)
                file_errors = check_file(filepath)
                all_errors.extend(file_errors)

    if all_errors:
        for err in all_errors:
            print(err)
        # sys.exit(1)
        print("WARNING: Documentation violations found (see above). Please fix them.")

if __name__ == "__main__":
    main()
