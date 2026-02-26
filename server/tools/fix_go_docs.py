# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import os
import re
import sys

# Regex patterns
FUNC_PATTERN = re.compile(r'^func\s+([A-Z][a-zA-Z0-9_]*)\s*\((.*?)\)\s*(\(.*\)|[a-zA-Z0-9_*\[\]\.]*)?\s*{?')
METHOD_PATTERN = re.compile(r'^func\s+\((.*?)\)\s+([A-Z][a-zA-Z0-9_]*)\s*\((.*?)\)\s*(\(.*\)|[a-zA-Z0-9_*\[\]\.]*)?\s*{?')
TYPE_PATTERN = re.compile(r'^type\s+([A-Z][a-zA-Z0-9_]*)\s+')
VAR_PATTERN = re.compile(r'^var\s+([A-Z][a-zA-Z0-9_]*)\s+')
CONST_PATTERN = re.compile(r'^const\s+([A-Z][a-zA-Z0-9_]*)\s+')

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
            names_str, type_name = parts
            names = [n.strip() for n in names_str.split(',')]
            for name in names:
                if name:
                    parsed.append((name, type_name))
        else:
            if len(parts) == 1:
                parsed.append(("_", parts[0]))
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
        if len(parts) == 2 and parts[0].replace(',','').isalnum():
             parsed.append((parts[0], parts[1]))
        else:
            parsed.append(("", r))
    return parsed

def parse_doc(doc_lines):
    sections = {
        "description": [],
        "summary": [],
        "parameters": [],
        "returns": [],
        "errors": [],
        "side_effects": []
    }

    current_section = "description"

    for line in doc_lines:
        clean = line.strip()
        if clean.startswith('//'):
            content = clean[2:].strip()

            if content.lower().startswith("summary:"):
                current_section = "summary"
                # content might be "Summary: The summary."
                # We want to keep the whole line or split?
                # Let's keep the content line as is, but mark section switch.
                # Actually, standard is:
                # Summary: ...
                sections[current_section].append(content)
                continue
            elif content.lower().startswith("parameters:"):
                current_section = "parameters"
                sections[current_section].append(content)
                continue
            elif content.lower().startswith("returns:"):
                current_section = "returns"
                sections[current_section].append(content)
                continue
            elif content.lower().startswith("errors:"):
                current_section = "errors"
                sections[current_section].append(content)
                continue
            elif content.lower().startswith("side effects:"):
                current_section = "side_effects"
                sections[current_section].append(content)
                continue

            # If empty line (just //), it might be separator.
            if not content:
                # separators are fine, keep them in current section if it's description?
                # or drop them?
                # Better to keep empty lines in description, but drop in others?
                if current_section == "description":
                    sections[current_section].append("")
                continue

            sections[current_section].append(content)

    return sections

def generate_doc(name, kind, params, returns, existing_doc):
    parsed = parse_doc(existing_doc)

    lines = []

    # Description
    desc = parsed["description"]
    if not desc:
        n = nice_name(name)
        if kind == "type":
             lines.append(f"// {name} represents a {n}.")
        elif kind in ("func", "method"):
            if name.startswith("New"):
                 lines.append(f"// {name} creates a new {nice_name(name[3:])}.")
            elif name.startswith("Get"):
                 lines.append(f"// {name} retrieves the {nice_name(name[3:])}.")
            else:
                 lines.append(f"// {name} {n}.")
        else:
             lines.append(f"// {name} is the {n}.")
    else:
        # Check if first line starts with Name
        if desc and not desc[0].startswith(name):
             # Prepend name if missing? Or just trust existing doc?
             # Standard says: "Calculates the tax" -> "CalculateTax calculates..."
             # Go convention usually requires function name at start.
             # Let's verify if the first word is the function name.
             first_word = desc[0].split(' ')[0]
             if first_word != name:
                  lines.append(f"// {name} {desc[0][0].lower() + desc[0][1:]}")
                  for l in desc[1:]: lines.append(f"// {l}")
             else:
                  for l in desc: lines.append(f"// {l}")
        else:
             for l in desc: lines.append(f"// {l}")

    lines.append("//")

    # Summary
    if parsed["summary"]:
        for l in parsed["summary"]:
            lines.append(f"// {l}")
    else:
        # Infer summary from first line of description
        # Extract content after Name
        first_desc = lines[0][3:] # strip //
        # remove name
        if first_desc.startswith(name + " "):
            s = first_desc[len(name)+1:].strip()
            # take first sentence
            s = s.split('.')[0] + "."
            lines.append(f"// Summary: {s}")
        else:
            lines.append(f"// Summary: {name} {nice_name(name)}.")

    # Parameters
    if kind in ("func", "method"):
        if parsed["parameters"]:
             lines.append("//")
             for l in parsed["parameters"]: lines.append(f"// {l}")
        elif params:
             lines.append("//")
             lines.append("// Parameters:")
             for pname, ptype in params:
                 desc = f"The {nice_name(pname)}."
                 if pname == "ctx": desc = "The context for the request."
                 if pname == "_": desc = "Ignored."
                 lines.append(f"//   - {pname} ({ptype}): {desc}")

    # Returns
    if kind in ("func", "method"):
        if parsed["returns"]:
             lines.append("//")
             for l in parsed["returns"]: lines.append(f"// {l}")
        elif returns:
             lines.append("//")
             lines.append("// Returns:")
             for rname, rtype in returns:
                 desc = "The result."
                 if rtype == "error": desc = "An error if the operation fails."
                 if rname:
                      lines.append(f"//   - {rname} ({rtype}): {desc}")
                 else:
                      lines.append(f"//   - {rtype}: {desc}")

    # Errors (optional, keep if exists)
    if parsed["errors"]:
        lines.append("//")
        for l in parsed["errors"]: lines.append(f"// {l}")

    # Side Effects
    if kind in ("func", "method"):
        if parsed["side_effects"]:
             lines.append("//")
             for l in parsed["side_effects"]: lines.append(f"// {l}")
        else:
             lines.append("//")
             lines.append("// Side Effects:")
             lines.append("//   - None.")

    return [l + "\n" for l in lines]

def process_file(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    new_lines = []
    comment_buffer = []

    i = 0
    while i < len(lines):
        line = lines[i]
        stripped = line.strip()

        if stripped.startswith('//'):
            if stripped.startswith('//go:') or stripped.startswith('// +build'):
                new_lines.extend(comment_buffer)
                comment_buffer = []
                new_lines.append(line)
            else:
                comment_buffer.append(line)
            i += 1
            continue

        m_func = FUNC_PATTERN.match(line)
        m_method = METHOD_PATTERN.match(line)
        m_type = TYPE_PATTERN.match(line)
        m_var = VAR_PATTERN.match(line)
        m_const = CONST_PATTERN.match(line)

        symbol = None
        kind = None
        params = []
        returns = []

        if m_func:
            symbol = m_func.group(1)
            kind = "func"
            params = parse_params(m_func.group(2))
            returns = parse_returns(m_func.group(3))
        elif m_method:
            symbol = m_method.group(2)
            kind = "method"
            params = parse_params(m_method.group(3))
            returns = parse_returns(m_method.group(4))
        elif m_type:
            symbol = m_type.group(1)
            kind = "type"
        elif m_var:
            symbol = m_var.group(1)
            kind = "var"
        elif m_const:
            symbol = m_const.group(1)
            kind = "const"

        if symbol:
            new_doc = generate_doc(symbol, kind, params, returns, comment_buffer)
            new_lines.extend(new_doc)
            comment_buffer = []
            new_lines.append(line)
        else:
            new_lines.extend(comment_buffer)
            comment_buffer = []
            new_lines.append(line)

        i += 1

    new_lines.extend(comment_buffer)

    with open(filepath, 'w') as f:
        f.writelines(new_lines)

def scan_dir(root_dir):
    for root, dirs, files in os.walk(root_dir):
        if "vendor" in root or "test" in root:
            continue

        for file in files:
            if file.endswith('.go') and not file.endswith('_test.go'):
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
                try:
                    process_file(arg)
                except Exception as e:
                    print(f"Error processing {arg}: {e}")
    else:
        scan_dir('server/pkg')
        scan_dir('server/cmd')
