import os
import re

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # The regex to match exported declarations (func, type, var, const)
    # Group 1: Preceding comments
    # Group 2: The type of declaration (func, func (recv), type, var, const)
    # Group 3: The name of the declaration
    # Group 4: The rest of the line
    pattern = re.compile(r'((?:^\s*//.*$\n)*)^(func(?:\s+\([^)]+\))?\s+|type\s+|var\s+|const\s+)([A-Z]\w*)(.*?)$', re.MULTILINE)

    def replacer(match):
        comments = match.group(1) or ""
        decl_type_full = match.group(2)
        name = match.group(3)
        rest = match.group(4)

        # Skip if it already has a summary
        if 'Summary:' in comments:
            return match.group(0)

        decl_type = decl_type_full.split()[0].strip()
        doc_lines = []

        # If there are NO preceding comments, we add the basic `// Name ...`
        if not comments:
            doc_lines.append(f"// {name} ...\n")
            if decl_type == 'func':
                doc_lines.append("//\n")
        elif not comments.endswith("//\n"):
             # Ensure a blank comment line before our structured block
             comments += "//\n"

        # Add the summary
        if decl_type == 'func':
            action = "Executes"
            if name.startswith('Get') or name.startswith('Read'): action = "Retrieves"
            elif name.startswith('Set') or name.startswith('Write'): action = "Updates"
            elif name.startswith('New') or name.startswith('Create'): action = "Initializes"
            elif name.startswith('Is') or name.startswith('Has'): action = "Checks"

            doc_lines.append(f"// Summary: {action} {name} operation.\n")
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
        else:
            doc_lines.append(f"// Summary: Represents a {name}.\n")

        new_comments = comments + "".join(doc_lines)
        return new_comments + decl_type_full + name + rest

    new_content = pattern.sub(replacer, content)

    # Process grouped var/const blocks: `const (...)` or `var (...)`
    lines = new_content.split('\n')
    in_block = False
    for i, line in enumerate(lines):
        if line.startswith('const (') or line.startswith('var ('):
            in_block = True
            continue
        if in_block and line.startswith(')'):
            in_block = False
            continue

        if in_block:
            # Check for exported block declarations e.g., `  MyConst = ...`
            match = re.match(r'^(\s+)([A-Z]\w*)\s*(?:=|type|[a-z])', line)
            if match:
                indent = match.group(1)
                name = match.group(2)

                # Check preceding lines for a summary
                j = i - 1
                has_summary = False
                while j >= 0 and lines[j].strip().startswith('//'):
                    if 'Summary:' in lines[j]:
                        has_summary = True
                        break
                    j -= 1

                if not has_summary:
                    # Modify the current line by prepending a summary comment
                    lines[i] = f"{indent}// Summary: Defines {name}.\n{line}"

    new_content = '\n'.join(lines)

    with open(filepath, 'w') as f:
        f.write(new_content)

import json
with open("missing_files.json", "r") as f:
    missing = json.load(f)

for filepath in missing:
    process_file(filepath)
