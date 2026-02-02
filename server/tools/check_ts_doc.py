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
    pattern_direct = re.compile(r'^\s*export\s+(default\s+)?(function|class|interface|type|const|enum)\s+([a-zA-Z0-9_]+)')

    # Regex for named exports block: export { Name, Name2 as Alias }
    pattern_named_block = re.compile(r'^\s*export\s+\{([^}]+)\}')

    missing_docs = []

    # Map of symbol name to definition line index
    symbol_defs = {}

    # 1. First pass: find all symbol definitions (const X =, function X, etc) to map them to lines
    # We only care about top-level definitions for this simple check
    def_pattern = re.compile(r'^\s*(?:export\s+)?(?:default\s+)?(function|class|interface|type|const|enum)\s+([a-zA-Z0-9_]+)')
    for i, line in enumerate(lines):
        match = def_pattern.match(line)
        if match:
            name = match.group(2)
            symbol_defs[name] = i

    # 2. Second pass: check direct exports
    for i, line in enumerate(lines):
        match = pattern_direct.match(line)
        if match:
            name = match.group(3)
            if not has_docstring(lines, i):
                missing_docs.append((i + 1, name))

        # Check named exports block
        match_block = pattern_named_block.match(line)
        if match_block:
            content_block = match_block.group(1)
            # Split by comma
            exports = [e.strip() for e in content_block.split(',')]
            for exp in exports:
                # Handle "Name as Alias"
                parts = re.split(r'\s+as\s+', exp)
                original_name = parts[0]

                # Check if we found the definition
                if original_name in symbol_defs:
                    def_line = symbol_defs[original_name]
                    if not has_docstring(lines, def_line):
                        missing_docs.append((def_line + 1, original_name))
                else:
                    # Maybe it's imported? If so, we skip documenting re-exports for now unless they are defined in this file.
                    pass

    return missing_docs

def has_docstring(lines, line_idx):
    """
    Checks if there is a JSDoc comment above the given line index.
    """
    j = line_idx - 1
    while j >= 0:
        prev = lines[j].strip()
        if not prev:
            j -= 1
            continue
        if prev.startswith('@'): # decorators
            j -= 1
            continue
        if prev.startswith('//'): # single line comment - NOT JSDoc
            j -= 1
            continue
        if prev.endswith('*/'):
            return True
        return False
    return False

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
