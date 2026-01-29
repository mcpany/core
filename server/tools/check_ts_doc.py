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

    # We want to catch:
    # export default function ...
    # export default class ...

    pattern = re.compile(r'^\s*export\s+(default\s+)?(function|class|interface|type|const|enum)\s+([a-zA-Z0-9_]+)')

    missing_docs = []

    for i, line in enumerate(lines):
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
                if prev.startswith('@'): # decorators
                    j -= 1
                    continue
                if prev.endswith('*/'):
                    has_doc = True
                break

            if not has_doc:
                missing_docs.append((i + 1, name))

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
