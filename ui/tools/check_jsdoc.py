# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import os
import re
import sys

def has_jsdoc(content, start_index):
    # Search backwards from start_index for the last non-whitespace characters
    # If they form the end of a comment block '*/', then check if it starts with '/**'

    i = start_index - 1
    while i >= 0 and content[i].isspace():
        i -= 1

    if i < 1:
        return False

    # Check for '*/'
    if content[i] == '/' and content[i-1] == '*':
        # Found end of comment, now find start
        j = i - 2
        while j >= 0:
            if content[j] == '/' and content[j+1] == '*' and content[j+2] == '*':
                # Found '/**'
                return True
            if content[j] == '/' and content[j+1] == '*' and content[j+2] != '*':
                 # Found '/*', but not '/**'. It's a regular comment.
                 # Stop here, as it breaks the continuity of JSDoc.
                 return False
            j -= 1
    return False

def check_file(filepath):
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
    except Exception as e:
        print(f"Error reading {filepath}: {e}")
        return []

    # Regex to find exported symbols
    patterns = [
        # export class/interface/type/enum Name
        re.compile(r'export\s+(?:default\s+)?(?:abstract\s+)?(class|interface|type|enum)\s+(\w+)'),
        # export function Name
        re.compile(r'export\s+(?:default\s+)?(?:async\s+)?function\s+(\w+)'),
        # export const/let/var Name = ...
        re.compile(r'export\s+(?:const|let|var)\s+(\w+)\s*='),
    ]

    missing = []

    for pattern in patterns:
        for match in pattern.finditer(content):
            start_index = match.start()
            name = match.group(match.lastindex) # The last group is the name

            if not has_jsdoc(content, start_index):
                line_num = content.count('\n', 0, start_index) + 1
                missing.append(f"{filepath}:{line_num}: missing doc for exported symbol '{name}'")

    return missing

def main():
    root_dir = 'ui/src'
    all_missing = []

    for root, dirs, files in os.walk(root_dir):
        for file in files:
            if file.endswith('.ts') or file.endswith('.tsx'):
                filepath = os.path.join(root, file)
                all_missing.extend(check_file(filepath))

    for m in all_missing:
        print(m)

    if all_missing:
        sys.exit(1)

if __name__ == '__main__':
    main()
