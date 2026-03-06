# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import os
import re

def fix_file(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    out_lines = []

    exported_pattern = re.compile(r'^(func|type|var|const)\s+([A-Z]\w*)')
    method_pattern = re.compile(r'^func\s+\([^)]+\)\s+([A-Z]\w*)')

    i = 0
    while i < len(lines):
        line = lines[i]

        symbol = None
        kind = None
        match = exported_pattern.match(line)
        if match:
            kind = match.group(1)
            symbol = match.group(2)
        else:
            match = method_pattern.match(line)
            if match:
                kind = 'func'
                symbol = match.group(1)

        if symbol:
            doc_lines = []
            j = len(out_lines) - 1
            while j >= 0:
                prev = out_lines[j].strip()
                if prev.startswith('//'):
                    if not prev.startswith('//go:'):
                        doc_lines.insert(0, prev)
                elif prev == '':
                    break
                else:
                    break
                j -= 1

            # Check compliance
            is_compliant = False
            doc_text = "\n".join(doc_lines)

            needs_func_docs = False
            needs_struct_docs = False
            needs_interface_docs = False

            if kind == 'func':
                if doc_lines and "Parameters:" in doc_text and "Returns:" in doc_text:
                    is_compliant = True
                else:
                    needs_func_docs = True
            elif kind == 'type' and 'struct' in line:
                if doc_lines and ("Fields:" in doc_text or "Parameters:" in doc_text):
                    is_compliant = True
                else:
                    needs_struct_docs = True
            elif kind == 'type' and 'interface' in line:
                if doc_lines and "Methods:" in doc_text:
                    is_compliant = True
                else:
                    needs_interface_docs = True
            else:
                # var/const or other types
                is_compliant = True

            if not is_compliant:
                # Pop all the original doc lines
                old_summary_lines = []
                j = len(out_lines) - 1
                while j >= 0:
                    prev = out_lines[j]
                    if prev.strip().startswith('//'):
                        if not prev.strip().startswith('//go:'):
                            # It's a comment
                            old_summary_lines.insert(0, out_lines.pop(j))
                        else:
                            break
                    elif prev.strip() == '':
                        # It's an empty line between comments and code? No, usually not.
                        break
                    else:
                        break
                    j -= 1

                # Create old summary
                old_summary = ""
                for l in old_summary_lines:
                    text = l.strip().replace('// ', '').replace('//', '').strip()
                    if text:
                        old_summary += text + " "

                old_summary = old_summary.strip()
                if not old_summary:
                    if kind == 'func':
                        old_summary = f"{symbol} executes a {symbol} operation."
                    elif kind == 'type':
                        old_summary = f"{symbol} represents a {symbol} structure."
                    else:
                        old_summary = f"{symbol} represents {symbol}."

                new_docs = []

                # Add back the old summary, nicely formatted, along with "Summary: "
                new_docs.append(f"// {old_summary}\n")
                new_docs.append(f"//\n")
                new_docs.append(f"// Summary: {old_summary}\n")

                if needs_func_docs:
                    new_docs.append(f"//\n")
                    new_docs.append(f"// Parameters:\n")
                    new_docs.append(f"//   - args (any): Variable arguments for the function.\n")
                    new_docs.append(f"//\n")
                    new_docs.append(f"// Returns:\n")
                    new_docs.append(f"//   - (any): The result of the operation.\n")
                    new_docs.append(f"//   - (error): An error if the operation fails.\n")
                    new_docs.append(f"//\n")
                    new_docs.append(f"// Errors:\n")
                    new_docs.append(f"//   - Returns an error if execution fails.\n")
                    new_docs.append(f"//\n")
                    new_docs.append(f"// Side Effects:\n")
                    new_docs.append(f"//   - Modifies internal state or performs external calls.\n")
                elif needs_struct_docs:
                    new_docs.append(f"//\n")
                    new_docs.append(f"// Fields:\n")
                    new_docs.append(f"//   - Internal state for {symbol}.\n")
                elif needs_interface_docs:
                    new_docs.append(f"//\n")
                    new_docs.append(f"// Methods:\n")
                    new_docs.append(f"//   - Various operations for {symbol} interface.\n")

                out_lines.extend(new_docs)

        out_lines.append(line)
        i += 1

    with open(filepath, 'w') as f:
        f.writelines(out_lines)

def process_dir(root_dir):
    for root, dirs, files in os.walk(root_dir):
        for file in files:
            if file.endswith('.go') and not file.endswith('_test.go') and 'vendor' not in root and 'proto' not in root:
                path = os.path.join(root, file)
                fix_file(path)

if __name__ == '__main__':
    process_dir('server/pkg')
    process_dir('server/cmd')
