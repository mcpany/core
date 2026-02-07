# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

"""
This script checks for missing documentation (JSDoc) on exported symbols in TypeScript files.
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
        A list of tuples containing (line_number, symbol_name) for missing docs.
    """
    with open(filepath, 'r') as f:
        content = f.read()

    lines = content.split('\n')

    # Regex for exported functions/classes/interfaces/types/consts
    # export function Name
    # export class Name
    # export interface Name
    # export type Name
    # export const Name
    pattern = re.compile(r'^\s*export\s+(default\s+)?(function|class|interface|type|const|enum)\s+([a-zA-Z0-9_]+)')

    # Regex for "export { ... }" blocks (naive check)
    export_block_pattern = re.compile(r'^\s*export\s+\{(.*)\}')

    missing_docs = []

    for i, line in enumerate(lines):
        # standard exports
        match = pattern.match(line)
        if match:
            name = match.group(3)
            # Check for docstring above
            has_doc = False
            j = i - 1
            while j >= 0:
                prev = lines[j].strip()
                if not prev:
                    j -= 1
                    continue
                if prev.startswith('@') or prev.startswith('//'): # decorators or comments
                    j -= 1
                    continue
                if prev.endswith('*/'):
                    has_doc = True
                break

            if not has_doc:
                missing_docs.append((i + 1, name))

        # export blocks
        block_match = export_block_pattern.match(line)
        if block_match:
            # For export { ... }, we need to check if the symbols are documented where they are defined.
            # This is hard to do with regex without parsing the whole file.
            # For now, we assume if they are defined in this file, they should be documented.
            # We can just warn about them or try to find the definition.
            # Let's try to find the definition.
            symbols = [s.strip() for s in block_match.group(1).split(',')]
            for sym in symbols:
                # Handle "foo as bar"
                if " as " in sym:
                    sym = sym.split(" as ")[0].strip()

                # Look for definition in the file
                # const sym = ...
                # function sym(...) ...
                # class sym ...
                # interface sym ...
                # type sym ...

                # Simple check: search for definition line
                def_pattern = re.compile(r'^\s*(const|function|class|interface|type|enum)\s+' + re.escape(sym) + r'\b')
                found_def = False
                for k, l in enumerate(lines):
                    if def_pattern.match(l):
                        found_def = True
                        # Check docs above definition
                        has_doc = False
                        m = k - 1
                        while m >= 0:
                            prev = lines[m].strip()
                            if not prev:
                                m -= 1
                                continue
                            if prev.startswith('@') or prev.startswith('//'):
                                m -= 1
                                continue
                            if prev.endswith('*/'):
                                has_doc = True
                            break

                        if not has_doc:
                             missing_docs.append((i + 1, f"{sym} (via export {{}} at line {i+1})"))
                        break

                # If not found definition, maybe it's imported? We ignore imports for now.

    return missing_docs

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
                    for line_num, name in missing:
                        print(f"{filepath}:{line_num}: missing doc for exported symbol {name}")

    if has_errors:
        sys.exit(1)

if __name__ == "__main__":
    main()
