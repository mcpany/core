# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

"""
This script checks for missing documentation (JSDoc) on exported symbols in TypeScript files.
"""

import os
import re
import sys

def get_doc_block(lines, export_line_idx):
    """
    Backtracks from export line to find JSDoc block.
    Returns the docstring content as a string, or None.
    """
    i = export_line_idx - 1
    # Skip decorators and empty lines
    while i >= 0:
        line = lines[i].strip()
        if not line or line.startswith('@'):
            i -= 1
            continue
        break

    if i < 0:
        return None

    if lines[i].strip().endswith('*/'):
        end_idx = i
        # Find start
        while i >= 0:
            if lines[i].strip().startswith('/**'):
                # Found block
                # Extract content
                doc_lines = []
                for k in range(i, end_idx + 1):
                    # Strip /**, */, *
                    l = lines[k].strip()
                    if l.startswith('/**'):
                        l = l[3:]
                    if l.endswith('*/'):
                        l = l[:-2]
                    l = l.strip()
                    if l.startswith('*'):
                        l = l[1:]
                    doc_lines.append(l.strip())
                return "\n".join(doc_lines)
            i -= 1
    return None

def check_structure(doc, name, symbol_type):
    issues = []
    if not doc or not doc.strip():
        return ["empty docstring"]

    # Check Summary (first line not empty)
    lines = doc.strip().split('\n')
    first_line = lines[0].strip()
    if not first_line:
        issues.append("missing summary")
    elif first_line.startswith("Copyright"):
         pass

    # Check Side Effects (Mandatory per AGENTS.md for logic components)
    # Skip for purely structural types
    if symbol_type not in ['interface', 'type', 'enum']:
        # For const, it might be a variable or function.
        # Ideally we'd check if it's a function, but for now we enforce it unless we are sure.
        # If it's a UPPER_CASE const, likely a value?
        if symbol_type == 'const' and name.isupper():
            pass # Skip side effects for constants like MAX_RETRIES
        elif "Side Effects:" not in doc:
            issues.append("missing 'Side Effects:' section")

    return issues

def check_file(filepath):
    """
    Checks a single file for missing documentation on exported symbols.
    """
    with open(filepath, 'r') as f:
        content = f.read()

    lines = content.split('\n')

    # Regex for exported functions/classes/interfaces/types/consts
    # Capture: 1=default(opt), 2=type, 3=name
    pattern = re.compile(r'^\s*export\s+(default\s+)?(function|class|interface|type|const|enum)\s+([a-zA-Z0-9_]+)')

    errors = []

    for i, line in enumerate(lines):
        match = pattern.match(line)
        if match:
            symbol_type = match.group(2)
            name = match.group(3)
            doc = get_doc_block(lines, i)

            if not doc:
                errors.append((i + 1, name, "missing docstring"))
            else:
                structure_issues = check_structure(doc, name, symbol_type)
                for issue in structure_issues:
                    errors.append((i + 1, name, issue))

    return errors

def process_path(path):
    has_errors = False
    if os.path.isfile(path):
        missing = check_file(path)
        if missing:
            has_errors = True
            for line_num, name, issue in missing:
                print(f"{path}:{line_num}: {issue} for exported symbol {name}")
    else:
        for dirpath, dirnames, filenames in os.walk(path):
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
                        for line_num, name, issue in missing:
                            print(f"{filepath}:{line_num}: {issue} for exported symbol {name}")
    return has_errors

def main():
    root_path = 'ui/src'
    if len(sys.argv) > 1:
        root_path = sys.argv[1]

    if process_path(root_path):
        sys.exit(1)

if __name__ == "__main__":
    main()
