import os
import re

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # Regex to find declarations with their preceding comments
    # Match: optional preceding comment lines, then the declaration line
    pattern = r'((?:^\s*//.*$\n)*)^(func\s+(?:(?:\([^)]+\)\s+)|)|type\s+|var\s+|const\s+)([A-Z]\w*)(.*?)$'

    def replacer(match):
        comments = match.group(1)
        decl_type_full = match.group(2)
        name = match.group(3)
        rest = match.group(4)

        # skip if 'Summary:' already in comments
        if 'Summary:' in comments:
            return match.group(0)

        # extract just the type: 'func', 'type', 'var', 'const'
        decl_type = decl_type_full.split()[0]

        doc_lines = []
        if not comments:
            doc_lines.append(f"// {name} ...\n")
            doc_lines.append("//\n")

        # Use descriptive summaries for types
        if decl_type == 'func':
            action = "Executes"
            if name.startswith('Get') or name.startswith('Read'): action = "Retrieves"
            elif name.startswith('Set') or name.startswith('Write'): action = "Updates"
            elif name.startswith('New') or name.startswith('Create'): action = "Initializes"
            elif name.startswith('Is') or name.startswith('Has'): action = "Checks"
            doc_lines.append(f"// Summary: {action} {name} operation.\n")
        else:
            doc_lines.append(f"// Summary: Represents a {name}.\n")

        if decl_type == 'func':
            doc_lines.append("//\n")
            doc_lines.append("// Parameters:\n")
            doc_lines.append("//   - TODO: Document parameters.\n")
            doc_lines.append("//\n")
            doc_lines.append("// Returns:\n")
            doc_lines.append("//   - TODO: Document returns.\n")
            doc_lines.append("//\n")
            doc_lines.append("// Errors:\n")
            doc_lines.append("//   - TODO: Document errors.\n")
            doc_lines.append("//\n")
            doc_lines.append("// Side Effects:\n")
            doc_lines.append("//   - None.\n")

        return comments + "".join(doc_lines) + decl_type_full + name + rest

    new_content = re.sub(pattern, replacer, content, flags=re.MULTILINE)

    # Check if there are unhandled block vars/consts
    # A block looks like `const (\n\t//...\n\tName = ...\n)`
    # This requires more complex logic to insert comments inside the block.
    # We will do a simple line-by-line replace for block items.

    in_block = False
    block_type = None
    lines = new_content.split('\n')
    for i, line in enumerate(lines):
        if line.startswith('const (') or line.startswith('var ('):
            in_block = True
            block_type = 'const' if line.startswith('const') else 'var'
            continue
        if in_block and line.startswith(')'):
            in_block = False
            continue

        if in_block:
            # Check if it's an exported var/const
            # Might be preceded by comments
            match = re.match(r'^\s+([A-Z]\w*)\s*(?:=|type|int|string|bool|float)', line)
            if match:
                name = match.group(1)
                # Check previous lines for summary
                j = i - 1
                has_summary = False
                while j >= 0 and lines[j].strip().startswith('//'):
                    if 'Summary:' in lines[j]:
                        has_summary = True
                        break
                    j -= 1

                if not has_summary:
                    # Insert a comment
                    lines[i] = f"\t// Summary: Defines {name}.\n" + line

    new_content = '\n'.join(lines)

    with open(filepath, 'w') as f:
        f.write(new_content)

import sys
process_file("server/pkg/resource/dynamic_resource.go")
