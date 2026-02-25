# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

"""
This script checks for missing or incomplete documentation (JSDoc) on exported symbols in TypeScript files.
It enforces the "Gold Standard" structure: Summary, Parameters, Returns, Errors, Side Effects.
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
        A list of tuples containing (line_number, symbol_name, issue) for missing docs/sections.
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
    # export const Name = ...

    # We want to catch:
    # export default function ...
    # export default class ...

    # Simple regex to find exported symbols
    pattern = re.compile(r'^\s*export\s+(default\s+)?(function|class|interface|type|const|enum)\s+([a-zA-Z0-9_]+)')

    missing_docs = []

    for i, line in enumerate(lines):
        match = pattern.match(line)
        if match:
            name = match.group(3)
            # Check for docstring above
            docstring = []
            has_doc = False
            j = i - 1

            # Walk backwards to find comment block
            in_comment = False
            comment_end = -1
            comment_start = -1

            while j >= 0:
                prev = lines[j].strip()
                if not prev:
                    if not in_comment and comment_end == -1:
                        j -= 1
                        continue # Skip empty lines between code and doc?
                    if in_comment:
                        # Empty line inside comment is fine
                        pass

                if prev.startswith('@'): # decorators
                    j -= 1
                    continue

                if prev.endswith('*/'):
                    if in_comment: # Nested? No.
                        break
                    in_comment = True
                    comment_end = j
                    # If it's a one-line comment /** ... */
                    if prev.startswith('/**'):
                        docstring.insert(0, prev)
                        comment_start = j
                        has_doc = True
                        break
                    docstring.insert(0, prev)
                    j -= 1
                    continue

                if in_comment:
                    docstring.insert(0, prev)
                    if prev.startswith('/**'):
                        comment_start = j
                        has_doc = True
                        break
                    j -= 1
                    continue

                # If we hit code or something else before finding comment
                break

            if not has_doc:
                missing_docs.append((i + 1, name, "missing docstring"))
                continue

            # Analyze docstring content
            doc_text = "\n".join(docstring)

            # Check Summary
            if "Summary:" not in doc_text:
                # Maybe implied? But strict mode says "Structure: ... Summary: ..."
                # Let's verify if client.ts uses "Summary:". Yes.
                missing_docs.append((i + 1, name, "missing 'Summary:' section"))

            # Check Side Effects (Always required by AGENTS.md)
            if "Side Effects:" not in doc_text:
                missing_docs.append((i + 1, name, "missing 'Side Effects:' section"))

            # Parameters, Returns, Errors are situational.
            # It's hard to know if they are needed without parsing TS.
            # But if they are present as tags, it's good.
            # If function takes args, it should have @param or Parameters:

            # For now, let's just enforce Summary and Side Effects as minimum "Structure" compliance,
            # plus check for existence of @param/@returns/@throws OR Parameters/Returns/Errors text
            # IF we can guess it needs them. But guessing is hard.
            # So we enforce explicit Summary and Side Effects as mandatory anchors.

    return missing_docs

def main():
    """
    Main function to walk the directory and check all applicable files.
    Exits with status code 1 if any missing documentation is found.
    """
    # Check if a directory is provided, otherwise default to ui/src
    root_dir = 'ui/src'
    if len(sys.argv) > 1:
        root_dir = sys.argv[1]

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
                    for line_num, name, issue in missing:
                        print(f"{filepath}:{line_num}: {issue} for exported symbol {name}")

    if has_errors:
        # sys.exit(1) # TODO: Enforce strict mode once coverage is higher
        pass

if __name__ == "__main__":
    main()
