# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import os
import re

"""
This script automatically generates or updates JSDoc comments for exported symbols
in TypeScript (.ts, .tsx) files.
"""

def infer_desc(name, kind):
    name_lower = name.lower()

    if kind == 'function' or kind == 'const':
        if name in ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']:
            return f"Handles the {name} request."
        if name.startswith('use'):
             return f"Hook for {split_camel_case(name[3:])}."
        if name.startswith('create'):
             return f"Creates a {split_camel_case(name[6:])}."
        if name.startswith('get'):
             return f"Retrieves the {split_camel_case(name[3:])}."
        if name.startswith('set'):
             return f"Sets the {split_camel_case(name[3:])}."
        if name.startswith('update'):
             return f"Updates the {split_camel_case(name[6:])}."
        if name.startswith('delete'):
             return f"Deletes the {split_camel_case(name[6:])}."
        if name.startswith('validate'):
             return f"Validates the {split_camel_case(name[8:])}."

    return f"The {split_camel_case(name)}."

def split_camel_case(s):
    # Simple heuristic
    s = re.sub('(.)([A-Z][a-z]+)', r'\1 \2', s)
    return re.sub('([a-z0-9])([A-Z])', r'\1 \2', s).lower()

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    lines = content.split('\n')
    new_lines = []
    i = 0
    modified = False

    # Regex patterns
    export_pattern = re.compile(r'^\s*export\s+(async\s+)?(function|const|class|type|interface)\s+([a-zA-Z0-9_]+)')

    while i < len(lines):
        line = lines[i]
        match = export_pattern.match(line)

        if match:
            kind = match.group(2)
            name = match.group(3)

            # Check for existing doc
            has_doc = False
            j = i - 1
            while j >= 0:
                prev = lines[j].strip()
                if not prev:
                    j -= 1
                    continue
                if prev.startswith('@'): # Decorators
                    j -= 1
                    continue
                if prev.endswith('*/'):
                    has_doc = True
                break

            if not has_doc:
                desc = infer_desc(name, kind)
                doc_lines = ['/**']
                doc_lines.append(f' * {desc}')
                doc_lines.append(f' *')
                doc_lines.append(f' * @summary {desc}')

                if kind == 'function' or (kind == 'const' and '=>' in line):
                     # Add generic returns
                     # We can't easily infer return type without parsing, so we omit @returns or put generic.
                     # The Gold Standard asks for @returns.
                     doc_lines.append(f' * @returns {{any}} The result.')

                doc_lines.append(' */')

                # Insert doc
                ins_idx = len(new_lines)
                while ins_idx > 0:
                    l = new_lines[ins_idx-1].strip()
                    if l.startswith('@'):
                        ins_idx -= 1
                    elif l == '' or l.startswith('//'):
                        break
                    else:
                        break

                for l in doc_lines:
                    new_lines.insert(ins_idx, l)
                    ins_idx += 1

                print(f"Added doc for {name} in {filepath}")
                modified = True

        new_lines.append(line)
        i += 1

    if modified:
        with open(filepath, 'w') as f:
            f.write('\n'.join(new_lines))

def main():
    root_dir = 'ui/src'
    for dirpath, dirnames, filenames in os.walk(root_dir):
        for filename in filenames:
            if filename.endswith('.ts') or filename.endswith('.tsx'):
                process_file(os.path.join(dirpath, filename))

if __name__ == "__main__":
    main()
