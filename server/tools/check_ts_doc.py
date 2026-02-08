# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

"""
This script checks for missing documentation (JSDoc) on exported symbols in TypeScript files.
It strictly checks direct exports and named exports, including multiline exports.
"""

import os
import re
import sys

def check_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    lines = content.split('\n')
    missing_docs = []

    definitions = {}

    # Definitions
    def_pattern = re.compile(r'^\s*(?:export\s+)?(?:default\s+)?(?:abstract\s+)?(?:async\s+)?(?:function|class|interface|type|enum)\s+([a-zA-Z0-9_]+)')
    const_pattern = re.compile(r'^\s*(?:export\s+)?(?:const|var|let)\s+([a-zA-Z0-9_]+)')

    for i, line in enumerate(lines):
        match = def_pattern.match(line)
        if match:
            definitions[match.group(1)] = i
            continue
        match = const_pattern.match(line)
        if match:
            definitions[match.group(1)] = i
            continue

    exported_names = set()

    # Direct exports
    export_pattern = re.compile(r'^\s*export\s+(?:default\s+)?(?:abstract\s+)?(?:async\s+)?(?:function|class|interface|type|enum)\s+([a-zA-Z0-9_]+)')
    export_const_pattern = re.compile(r'^\s*export\s+(?:const|var|let)\s+([a-zA-Z0-9_]+)')

    for i, line in enumerate(lines):
        match = export_pattern.match(line)
        if match:
            exported_names.add(match.group(1))
            continue
        match = export_const_pattern.match(line)
        if match:
            exported_names.add(match.group(1))
            continue

    # Multiline named exports
    export_block_regex = re.compile(r'export\s*\{([^}]+)\}', re.DOTALL)

    for match in export_block_regex.finditer(content):
        # We need to see if 'from' is immediately after '}'
        # This is tricky because content might be huge and we iterate all matches.

        # Check suffix
        suffix = content[match.end():]
        # Skip whitespace/comments
        suffix_clean = re.sub(r'^\s+', '', suffix)
        if suffix_clean.startswith('from'):
            continue

        block_content = match.group(1)
        # Remove comments inside block
        block_content = re.sub(r'//.*', '', block_content)
        block_content = re.sub(r'/\*.*?\*/', '', block_content, flags=re.DOTALL)

        # Replace newlines with comma to split easily
        block_content = block_content.replace('\n', ',')

        items = [x.strip() for x in block_content.split(',')]
        for item in items:
            if not item: continue
            parts = item.split(' as ')
            original_name = parts[0].strip()
            exported_names.add(original_name)

    # Check docs
    for name in exported_names:
        if name not in definitions:
            continue

        def_line = definitions[name]
        has_doc = False
        j = def_line - 1
        while j >= 0:
            prev = lines[j].strip()
            if not prev:
                j -= 1
                continue
            if prev.startswith('// eslint-') or prev.startswith('@'):
                j -= 1
                continue

            if prev.endswith('*/'):
                has_doc = True
            break

        if not has_doc:
            missing_docs.append((def_line + 1, name))

    return missing_docs

def main():
    root_dir = 'ui/src'
    # Allow overriding root dir via args
    if len(sys.argv) > 1:
        root_dir = sys.argv[1]

    has_errors = False

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
                    has_errors = True
                    for line, name in missing:
                        print(f"{filepath}:{line}: missing doc for exported symbol {name}")

    if has_errors:
        sys.exit(1)

if __name__ == "__main__":
    main()
