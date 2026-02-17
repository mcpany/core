# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

"""
This script checks for missing or non-compliant documentation (JSDoc) on exported symbols in TypeScript files.
"""

import os
import re
import sys

strict_mode = os.environ.get("STRICT_DOC_CHECK") == "true"

def get_docstring(lines, index):
    """
    Extracts the docstring ending at lines[index].
    Returns (docstring, start_line_index) or (None, -1).
    """
    if index < 0:
        return None, -1

    line = lines[index].strip()
    if not line.endswith('*/'):
        return None, -1

    doc_lines = []
    i = index
    while i >= 0:
        l = lines[i].strip()
        doc_lines.insert(0, l)
        if l.startswith('/**'):
            return "\n".join(doc_lines), i
        if l.startswith('/*') and not l.startswith('/**'): # valid block comment but not JSDoc? allow it? JSDoc starts with /**
             # Strict JSDoc requires /**
             return None, -1
        i -= 1

    return None, -1

def count_args(signature):
    """
    Counts arguments in a function signature.
    Heuristic: count commas at top level of parentheses.
    """
    # Remove nested parens/braces/brackets to avoid confusion
    # This is a simple heuristic.
    # We only care if there are args or not.

    # Check if inside (...)
    match = re.search(r'\((.*)\)', signature)
    if not match:
        return 0

    args_str = match.group(1).strip()
    if not args_str:
        return 0

    # If args_str is just space, return 0
    if not args_str:
        return 0

    return 1 # Assume at least one if string not empty?
    # Better: split by comma, respecting nesting?
    # For now, just checking presence of @param if args exist is enough?
    # Let's count commas.
    return args_str.count(',') + 1

def check_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    lines = content.split('\n')
    missing_docs = []

    # Regex for exported functions
    # export function Name(...) ...
    # export const Name = (...) => ...
    # export const Name = function(...) ...

    # We iterate line by line to handle context
    for i, line in enumerate(lines):
        line = line.strip()

        is_exported_func = False
        func_name = ""
        has_args = False
        has_return = False

        # Check for export function
        match = re.match(r'^export\s+(default\s+)?function\s+([a-zA-Z0-9_]+)\s*\(', line)
        if match:
            is_exported_func = True
            func_name = match.group(2)
            # Check args in this line? Or subsequent?
            # Assuming single line signature for now or simple multiline
            if '()' not in line:
                has_args = True # Likely has args if not explicitly empty ()

            # Check return type
            if '): void' in line or '): Promise<void>' in line:
                has_return = False
            elif '):' in line:
                has_return = True

        # Check for export const arrow func
        match = re.match(r'^export\s+const\s+([a-zA-Z0-9_]+)\s*=\s*(\(|async\s*\()', line)
        if match:
            is_exported_func = True
            func_name = match.group(1)
            if '()' not in line and 'async ()' not in line:
                has_args = True

            if '): void' in line or '): Promise<void>' in line:
                has_return = False
            elif '):' in line:
                has_return = True

        if is_exported_func:
            # Check for docstring
            # Look at previous line
            prev_idx = i - 1
            while prev_idx >= 0 and (not lines[prev_idx].strip() or lines[prev_idx].strip().startswith('@')):
                prev_idx -= 1

            doc, _ = get_docstring(lines, prev_idx)

            if not doc:
                missing_docs.append((i + 1, func_name, "Missing doc", True))
            else:
                # Check structure
                if has_args and '@param' not in doc:
                     missing_docs.append((i + 1, func_name, "Missing @param", False))

                if has_return and '@returns' not in doc and '@return' not in doc:
                     missing_docs.append((i + 1, func_name, "Missing @returns", False))

    return missing_docs

def main():
    root_dir = 'ui/src'
    has_errors = False

    if len(sys.argv) > 1:
        root_dir = sys.argv[1]

    for dirpath, dirnames, filenames in os.walk(root_dir):
        if 'node_modules' in dirnames:
            dirnames.remove('node_modules')

        for filename in filenames:
            if filename.endswith('.ts') or filename.endswith('.tsx'):
                if filename.endswith('.d.ts') or filename.endswith('.test.ts') or filename.endswith('.test.tsx'):
                    continue

                filepath = os.path.join(dirpath, filename)
                missing = check_file(filepath)

                if missing:
                    for line_num, name, reason, is_missing in missing:
                        if is_missing or strict_mode:
                            print(f"{filepath}:{line_num}: {reason} for exported symbol {name}")
                            has_errors = True

    if has_errors:
        sys.exit(1)

if __name__ == "__main__":
    main()
