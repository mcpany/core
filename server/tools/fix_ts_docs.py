# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import os
import re

def get_param_desc(name):
    name = name.strip()
    if name.startswith('_'): name = name[1:]
    if name == 'props': return "The component props."
    if name == 'children': return "The child elements."
    if name == 'className': return "The CSS class name."
    if name == 'ref': return "The reference."
    if name == 'key': return "The unique key."
    if 'id' in name.lower(): return f"The {name} identifier."
    if 'name' in name.lower(): return f"The {name} value."
    if 'callback' in name.lower() or name.startswith('on'): return f"The {name} callback."
    return f"The {name} parameter."

def generate_doc_lines(name, type_kind, params):
    lines = []

    # Description
    desc = f"{name} {type_kind}."
    # Improve description based on name
    if name.startswith("use"):
        # Split camel case
        parts = re.findall(r'[A-Z][a-z]*', name)
        desc_parts = [p.lower() for p in parts]
        if not desc_parts and len(name) > 3: # usage -> use age
             desc_parts = [name[3:].lower()]

        desc = f"Hook to {name[3:].lower()}." # Fallback
        if desc_parts:
             desc = f"Hook to {' '.join(desc_parts)}."
    elif type_kind == 'component':
        desc = f"{name} component."
    elif type_kind == 'interface':
        desc = f"{name} interface."
    elif type_kind == 'type':
        desc = f"{name} type definition."

    lines.append(f" * {desc}")

    # Params
    for p in params:
        lines.append(f" * @param {p} - {get_param_desc(p)}")

    # Return
    if type_kind in ['function', 'method', 'hook']:
        if type_kind != 'component':
            lines.append(f" * @returns The result of {name}.")
        elif type_kind == 'component':
            lines.append(" * @returns The rendered component.")

    return lines

def parse_docstring(lines):
    # lines are the lines of the docstring excluding /** and */ markers
    desc = []
    params = {} # name -> desc
    returns = None

    content = []
    for line in lines:
        line = line.strip()
        if line.startswith('/**'): line = line[3:]
        if line.endswith('*/'): line = line[:-2]
        line = line.lstrip('*').strip()
        if line: content.append(line)

    for line in content:
        if line.startswith('@param'):
            match = re.match(r'@param\s+({[^}]+}|\[[^\]]+\]|[a-zA-Z0-9_\.]+)\s*(?:-\s*)?(.*)', line)
            if match:
                p_name = match.group(1).strip()
                p_desc = match.group(2).strip()
                params[p_name] = p_desc
        elif line.startswith('@returns') or line.startswith('@return'):
            match = re.match(r'@returns?\s+(.*)', line)
            if match:
                returns = match.group(1).strip()
        elif line.startswith('@'):
            pass
        else:
            desc.append(line)

    return desc, params, returns

def extract_params(params_str):
    if not params_str: return []

    # Simple iteration to split by comma respecting braces
    depth = 0
    current = ""
    args = []
    for char in params_str:
        if char == ',' and depth == 0:
            args.append(current)
            current = ""
        else:
            if char in '{[<(': depth += 1
            if char in '}]>)' and depth > 0: depth -= 1
            current += char
    if current: args.append(current)

    final_params = []
    for arg in args:
        arg = arg.strip()
        if not arg: continue

        # Remove type annotation
        arg_depth = 0
        split_idx = -1
        for idx, char in enumerate(arg):
             if char in '{[<(': arg_depth += 1
             if char in '}]>)' and arg_depth > 0: arg_depth -= 1
             if char == ':' and arg_depth == 0:
                 split_idx = idx
                 break

        if split_idx != -1:
            arg_name = arg[:split_idx].strip()
        else:
            arg_name = arg.strip()

        # Remove default value
        if '=' in arg_name:
            arg_name = arg_name.split('=')[0].strip()

        # Check if destructuring
        if arg_name.startswith('{') and arg_name.endswith('}'):
             inner = arg_name[1:-1]
             # Recursively split inner props
             # Inner props usually don't have types like outer args, but they might have renaming a:b
             # We can assume comma split is safe enough for destructuring usually (no types inside usually)
             sub_props = inner.split(',')
             for sp in sub_props:
                 sp = sp.strip()
                 # handle renaming: a: b
                 if ':' in sp:
                     sp = sp.split(':')[0].strip()

                 # Remove default { a = 1 }
                 if '=' in sp:
                     sp = sp.split('=')[0].strip()

                 if sp and sp != '...':
                     final_params.append(sp)
        else:
            if arg_name and arg_name != '...':
                final_params.append(arg_name)

    return final_params

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    lines = content.split('\n')
    new_lines = []
    i = 0
    modified = False

    # Regex patterns
    func_pattern = re.compile(r'^\s*export\s+(async\s+)?function\s+([a-zA-Z0-9_]+)\s*\(([^)]*)\)')
    arrow_pattern = re.compile(r'^\s*export\s+const\s+([a-zA-Z0-9_]+)\s*=\s*(async\s*)?(\([^)]*\)|[a-zA-Z0-9_]+)\s*=>')
    hoc_pattern = re.compile(r'^\s*export\s+const\s+([a-zA-Z0-9_]+)\s*=\s*(memo|forwardRef|dynamic)\(')
    class_pattern = re.compile(r'^\s*export\s+class\s+([a-zA-Z0-9_]+)')
    interface_pattern = re.compile(r'^\s*export\s+interface\s+([a-zA-Z0-9_]+)')
    type_pattern = re.compile(r'^\s*export\s+type\s+([a-zA-Z0-9_]+)')

    while i < len(lines):
        line = lines[i]

        # Check if doc exists
        has_doc = False
        doc_start = -1
        doc_end = -1

        j = i - 1
        while j >= 0:
            prev = lines[j].strip()
            if not prev:
                j -= 1
                continue
            if prev.startswith('@'): # decorators
                j -= 1
                continue
            if prev.endswith('*/'):
                has_doc = True
                doc_end = j
                # Find start
                k = j
                while k >= 0:
                    if '/**' in lines[k]:
                        doc_start = k
                        break
                    k -= 1
            break

        def insert_new_doc(name, kind, params):
            nonlocal modified, new_lines
            doc_body = generate_doc_lines(name, kind, params)
            doc_full = ["/**"] + doc_body + [" */"]

            insert_pos = len(new_lines)
            while insert_pos > 0 and new_lines[insert_pos-1].strip().startswith('@'):
                insert_pos -= 1
            for l in doc_full:
                new_lines.insert(insert_pos, l)
                insert_pos += 1
            modified = True

        def update_existing_doc(name, kind, params):
            nonlocal modified, new_lines, doc_start, doc_end

            current_doc_lines = lines[doc_start : doc_end+1]
            desc, existing_params, existing_returns = parse_docstring(current_doc_lines)

            # Clean junk params
            existing_keys = list(existing_params.keys())
            for k in existing_keys:
                if '{' in k or '}' in k or len(k) < 1:
                    del existing_params[k]

            missing_params = []
            for p in params:
                 if p not in existing_params:
                     missing_params.append(p)

            missing_return = False
            if kind in ['function', 'method', 'hook', 'component'] and not existing_returns:
                  if kind == 'component':
                      missing_return = True
                  elif kind == 'function' and not name.startswith('set'):
                      missing_return = True
                  elif kind == 'hook':
                      missing_return = True

            if missing_params or missing_return:
                new_doc = ["/**"]
                for d in desc:
                    if d: new_doc.append(f" * {d}")

                processed_params = set()
                for p in params:
                    if p in existing_params:
                        new_doc.append(f" * @param {p} - {existing_params[p]}")
                    else:
                        new_doc.append(f" * @param {p} - {get_param_desc(p)}")
                    processed_params.add(p)

                for p, d in existing_params.items():
                    if p not in processed_params:
                        new_doc.append(f" * @param {p} - {d}")

                if existing_returns:
                    new_doc.append(f" * @returns {existing_returns}")
                elif missing_return:
                    if kind == 'component':
                        new_doc.append(" * @returns The rendered component.")
                    else:
                        new_doc.append(f" * @returns The result of {name}.")

                new_doc.append(" */")

                # Replace in new_lines
                doc_len = doc_end - doc_start + 1
                decorators_count = 0
                for k in range(doc_end + 1, i):
                    if lines[k].strip().startswith('@'):
                        decorators_count += 1

                end_idx = len(new_lines) - decorators_count
                start_idx = end_idx - doc_len

                if start_idx < 0 or end_idx > len(new_lines) or new_lines[start_idx].strip() != lines[doc_start].strip():
                     print(f"Warning: Index mismatch for {name} in {filepath}. Skipping update.")
                     return

                new_lines[start_idx:end_idx] = new_doc
                modified = True
                print(f"Updated doc for {name} in {filepath}")


        match = func_pattern.match(line)
        if match:
            name = match.group(2)
            params_str = match.group(3)
            params = extract_params(params_str)
            kind = 'function'
            if name[0].isupper(): kind = 'component'
            if name.startswith('use'): kind = 'hook'

            if not has_doc: insert_new_doc(name, kind, params)
            else: update_existing_doc(name, kind, params)

        match = arrow_pattern.match(line)
        if match:
            name = match.group(1)
            params_str = match.group(3)
            if params_str.startswith('('): params_str = params_str[1:-1]
            params = extract_params(params_str)
            kind = 'function'
            if name[0].isupper(): kind = 'component'
            if name.startswith('use'): kind = 'hook'

            if not has_doc: insert_new_doc(name, kind, params)
            else: update_existing_doc(name, kind, params)

        match = hoc_pattern.match(line)
        if match:
            name = match.group(1)
            kind = 'component'
            params = ['props']
            if not has_doc: insert_new_doc(name, kind, params)
            else: update_existing_doc(name, kind, params)

        match = class_pattern.match(line)
        if match:
            name = match.group(1)
            if not has_doc: insert_new_doc(name, 'class', [])
            else: update_existing_doc(name, 'class', [])

        match = interface_pattern.match(line)
        if match:
            name = match.group(1)
            if not has_doc: insert_new_doc(name, 'interface', [])

        match = type_pattern.match(line)
        if match:
            name = match.group(1)
            if not has_doc: insert_new_doc(name, 'type', [])

        new_lines.append(line)
        i += 1

    if modified:
        with open(filepath, 'w') as f:
            f.write('\n'.join(new_lines))

def main():
    root_dir = 'ui/src'
    for dirpath, dirnames, filenames in os.walk(root_dir):
        for filename in filenames:
            if filename.endswith('.tsx') or filename.endswith('.ts'):
                if 'node_modules' in dirpath: continue
                if filename.endswith('.d.ts'): continue

                process_file(os.path.join(dirpath, filename))

if __name__ == "__main__":
    main()
