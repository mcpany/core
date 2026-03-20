# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import os
import re
import sys

# Regex patterns
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
    return [p for p in parsed if p and p != "{"]

def generate_doc(name, params, returns, receiver=None, is_type=False):
    nice = nice_name(name)
    summary = f"Represents a {name}." if is_type else f"Executes {name} operation."

    if name.startswith("New"):
        summary = f"Initializes a new {name[3:]}."
    elif name.startswith("Get"):
        summary = f"Retrieves {nice_name(name[3:])}."
    elif name.startswith("List"):
        summary = f"Retrieves a list of {nice_name(name[4:])}."

    lines = []
    lines.append(f"// {name} documentation.\n")
    lines.append("//\n")
    lines.append(f"// Summary: {summary}\n")

    if not is_type:
        if params:
            lines.append("//\n")
            lines.append("// Parameters:\n")
            for pname, ptype in params:
                lines.append(f"//   - {pname} ({ptype}): The {pname} parameter.\n")

        if returns:
            lines.append("//\n")
            lines.append("// Returns:\n")
            for rtype in returns:
                lines.append(f"//   - {rtype}: The resulting {rtype}.\n")
    return lines

def process_file(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()
    final_lines = []
    i = 0
    while i < len(lines):
        line = lines[i]
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
            # Check if there is already a comment
            has_existing = False
            if i > 0 and lines[i-1].strip().startswith('//'):
                has_existing = True

            if not has_existing:
                new_doc = generate_doc(target_name, params, returns, receiver, is_type)
                final_lines.extend(new_doc)
            elif i > 0:
                 # Check if existing doc has Summary:
                 doc_block = []
                 j = i - 1
                 while j >= 0 and lines[j].strip().startswith('//'):
                     doc_block.insert(0, lines[j])
                     j -= 1

                 needs_summary = True
                 needs_params = bool(params)
                 needs_returns = bool(returns)

                 for dline in doc_block:
                     if "Summary:" in dline: needs_summary = False
                     if "Parameters:" in dline: needs_params = False
                     if "Returns:" in dline: needs_returns = False

                 if needs_summary or needs_params or needs_returns:
                     # Append to existing doc
                     if needs_summary:
                         final_lines.append(f"// Summary: {target_name} operation.\n")
                     if needs_params:
                         final_lines.append("// Parameters:\n")
                         for pname, ptype in params:
                             final_lines.append(f"//   - {pname} ({ptype}): The {pname} parameter.\n")
                     if needs_returns:
                         final_lines.append("// Returns:\n")
                         for rtype in returns:
                             final_lines.append(f"//   - {rtype}: The resulting {rtype}.\n")
        final_lines.append(line)
        i += 1
    with open(filepath, 'w') as f:
        f.writelines(final_lines)

def scan_dir(root_dir):
    for root, dirs, files in os.walk(root_dir):
        for file in files:
            if file.endswith('.go') and not file.endswith('_test.go') and 'vendor' not in root:
                process_file(os.path.join(root, file))

if __name__ == '__main__':
    if len(sys.argv) > 1:
        for arg in sys.argv[1:]:
            if os.path.isdir(arg): scan_dir(arg)
            else: process_file(arg)
