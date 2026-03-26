# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import os
import re
import sys

# Regex patterns for exported symbols
FUNC_PATTERN = re.compile(r'^func\s+([A-Z][a-zA-Z0-9_]*)\s*\((.*?)\)\s*(\(.*\)|[a-zA-Z0-9_*\[\]\.]*)?\s*{')
TYPE_PATTERN = re.compile(r'^type\s+([A-Z][a-zA-Z0-9_]*)\s+(struct|interface)\s*{')
METHOD_PATTERN = re.compile(r'^func\s+\((.*?)\)\s+([A-Z][a-zA-Z0-9_]*)\s*\((.*?)\)\s*(\(.*\)|[a-zA-Z0-9_*\[\]\.]*)?\s*{')

def nice_name(name):
    s1 = re.sub('(.)([A-Z][a-z]+)', r'\1 \2', name)
    s2 = re.sub('([a-z0-9])([A-Z])', r'\1 \2', s1)
    return s2.lower()

def parse_params(param_str):
    if not param_str:
        return []
    params = []
    current = ""
    depth = 0
    for char in param_str:
        if char == ',' and depth == 0:
            params.append(current.strip())
            current = ""
        else:
            if char in '([{': depth += 1
            elif char in ')]}': depth -= 1
            current += char
    if current:
        params.append(current.strip())

    parsed = []
    for p in params:
        parts = p.rsplit(' ', 1)
        if len(parts) == 2:
            names, type_name = parts
            for name in names.split(','):
                parsed.append((name.strip(), type_name.strip()))
        else:
            parsed.append((p, "unknown"))
    return parsed

def parse_returns(return_str):
    if not return_str:
        return []
    return_str = return_str.strip()
    if return_str.startswith('(') and return_str.endswith(')'):
        return_str = return_str[1:-1]
    returns = []
    current = ""
    depth = 0
    for char in return_str:
        if char == ',' and depth == 0:
            returns.append(current.strip())
            current = ""
        else:
            if char in '([{': depth += 1
            elif char in ')]}': depth -= 1
            current += char
    if current:
        returns.append(current.strip())
    parsed = []
    for r in returns:
        parts = r.rsplit(' ', 1)
        if len(parts) == 2 and '.' not in parts[0]:
             parsed.append(parts[1])
        else:
            parsed.append(r)
    return parsed

def generate_doc(name, params, returns, receiver=None, is_type=False):
    nice = nice_name(name)
    summary = f"{name} {nice}."

    if is_type:
        summary = f"Represents a {nice}."
    elif name.startswith("New"):
        obj = nice_name(name[3:])
        summary = f"Creates a new {obj}."
    elif name.startswith("Get"):
        obj = nice_name(name[3:])
        summary = f"Retrieves the {obj}."
    elif name.startswith("List"):
        obj = nice_name(name[4:])
        summary = f"Retrieves a list of {obj}."
    elif name.startswith("Save") or name.startswith("Create"):
        obj = nice_name(name[4:]) if name.startswith("Save") else nice_name(name[6:])
        summary = f"Persists the {obj}."
    elif name.startswith("Delete"):
        obj = nice_name(name[6:])
        summary = f"Deletes the {obj}."
    elif name.startswith("Update"):
        obj = nice_name(name[6:])
        summary = f"Updates the {obj}."
    elif name == "Execute":
        summary = "Executes the operation."

    lines = []
    lines.append(f"// {name} {summary[0].lower() + summary[1:]}\n")
    lines.append("//\n")
    lines.append(f"// Summary: {summary}\n".

    if not is_type:
        lines.append("//\n")
        lines.append("// Parameters:\n")
        if params:
            for pname, ptype in params:
                if pname == "_":
                    desc = "Unused parameter."
                elif pname == "ctx":
                    desc = "The context for the request."
                elif pname == "id":
                    desc = "The unique identifier."
                else:
                    desc = f"The {nice_name(pname)}."
                lines.append(f"//   - {pname} ({ptype}): {desc}\n")
        else:
            lines.append("//   - None.\n")

        lines.append("//\n")
        lines.append("// Returns:\n")
        if returns:
            for rtype in returns:
                desc = "The result."
                if rtype == "error": desc = "An error if the operation fails."
                lines.append(f"//   - {rtype}: {desc}\n")
        else:
             lines.append("//   - None.\n")

        lines.append("//\n")
        lines.append("// Errors:\n")
        if returns and "error" in returns:
            lines.append("//   - Returns an error if the operation fails.\n")
        else:
            lines.append("//   - None.\n")

        lines.append("//\n")
        lines.append("// Side Effects:\n")
        lines.append("//   - None.\n")

    return lines

def process_file(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    final_lines = []
    comment_buffer = []

    for i, line in enumerate(lines):
        stripped = line.strip()
        if stripped.startswith('//'):
            if stripped.startswith('//go:'):
                final_lines.extend(comment_buffer)
                comment_buffer = []
                final_lines.append(line)
            else:
                comment_buffer.append(line)
        else:
            match_func = FUNC_PATTERN.match(line)
            match_method = METHOD_PATTERN.match(line)
            match_type = TYPE_PATTERN.match(line)

            target_name = None
            params = []
            returns = []
            receiver = None
            is_type = False

            if match_func:
                target_name = match_func.group(1)
                params = parse_params(match_func.group(2))
                returns = parse_returns(match_func.group(3))
            elif match_method:
                receiver = match_method.group(1)
                target_name = match_method.group(2)
                params = parse_params(match_method.group(3))
                returns = parse_returns(match_method.group(4))
            elif match_type:
                target_name = match_type.group(1)
                is_type = True

            if target_name:
                has_comments = len(comment_buffer) > 0

                # If it has comments, let's see if they are already compliant
                is_compliant = False
                if has_comments:
                    text = "".join(comment_buffer)
                    if "Summary:" in text and "Parameters:" in text and "Returns:" in text:
                        is_compliant = True

                if is_compliant:
                    final_lines.extend(comment_buffer)
                else:
                    # Not compliant or no comments, generate new one
                    # We might want to merge, but for speed, let's overwrite if not compliant
                    new_doc = generate_doc(target_name, params, returns, receiver, is_type)
                    final_lines.extend(new_doc)

                comment_buffer = []
                final_lines.append(line)
            else:
                if comment_buffer:
                    final_lines.extend(comment_buffer)
                    comment_buffer = []
                final_lines.append(line)

    if comment_buffer:
        final_lines.extend(comment_buffer)

    with open(filepath, 'w') as f:
        f.writelines(final_lines)

def scan_dir(root_dir):
    for root, dirs, files in os.walk(root_dir):
        if any(x in root for x in ["vendor", "node_modules", "zz_generated"]):
            continue
        for file in files:
            if file.endswith('.go') and not any(file.endswith(x) for x in ['_test.go', '.pb.go', '.pb.gw.go']):
                path = os.path.join(root, file)
                try:
                    process_file(path)
                except Exception as e:
                    print(f"Error processing {path}: {e}")

if __name__ == '__main__':
    if len(sys.argv) > 1:
        for arg in sys.argv[1:]:
            if os.path.isdir(arg):
                scan_dir(arg)
            else:
                process_file(arg)
    else:
        scan_dir('.')
