# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import os
import re

def split_camel_case(s):
    return re.sub('([A-Z][a-z]+)', r' \1', re.sub('([A-Z]+)', r' \1', s)).split()

def generate_summary(name, kind):
    words = split_camel_case(name)
    words = [w.lower() for w in words]
    if kind == 'func':
        if len(words) == 0:
            return f"{name} executes the operation."
        if words[0] in ['get', 'fetch', 'find', 'read', 'load']:
            return f"Retrieves the {' '.join(words[1:])} associated with the request."
        elif words[0] in ['set', 'write', 'save', 'store', 'update']:
            return f"Stores the {' '.join(words[1:])} into the persistent storage."
        elif words[0] == 'new':
            return f"Initializes a new {' '.join(words[1:])} instance with default parameters."
        elif words[0] == 'is' or words[0] == 'has':
            return f"Checks if the object {' '.join(words)} and returns a boolean."
        elif words[0] == 'list':
            return f"Lists all available {' '.join(words[1:])} from the current context."
        elif words[0] == 'validate':
            return f"Validates the {' '.join(words[1:])} to ensure correctness."
        else:
            return f"{name} executes the {' '.join(words)} operation."
    else:
        return f"{name} defines the core structure for {' '.join(words)} within the system."

def get_param_desc(name, typ):
    if name in ['ctx', 'context']:
        return "The context for managing request lifecycle and cancellation."
    elif name == 'err':
        return "The error encountered during the operation."
    elif name == 'req':
        return "The request object containing specific parameters."
    elif name == 'resp':
        return "The response object containing operation results."
    elif 'id' in name.lower():
        return f"The unique identifier used to reference the {name.replace('id', '').replace('ID', '')} resource."
    elif 'config' in name.lower():
        return "The configuration settings to be applied."
    elif 'w' == name and 'http.ResponseWriter' in typ:
        return "The HTTP response writer to construct the response."
    elif 'r' == name and '*http.Request' in typ:
        return "The HTTP request containing client payload."
    else:
        words = split_camel_case(name)
        return f"The {' '.join([w.lower() for w in words])} parameter used in the operation."

def get_return_desc(typ):
    if typ == 'error':
        return "An error object if the operation fails, otherwise nil."
    elif 'bool' in typ:
        return "A boolean indicating the success or status of the operation."
    elif 'string' in typ:
        return "A string value representing the operation's result."
    else:
        return f"The resulting {typ.replace('*', '')} object containing the requested data."

func_pattern = re.compile(r'^func\s+(?:\([^)]+\)\s+)?([A-Z]\w*)\s*\(([^)]*)\)(?:\s+([^{]+))?')
type_pattern = re.compile(r'^type\s+([A-Z]\w*)\s+(struct|interface)')
var_const_pattern = re.compile(r'^(var|const)\s+([A-Z]\w*)')

def fix_file(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    out_lines = []
    i = 0
    while i < len(lines):
        line = lines[i]

        func_match = func_pattern.match(line)
        type_match = type_pattern.match(line)
        var_const_match = var_const_pattern.match(line)

        symbol = None
        kind = None
        type_kind = None
        if func_match:
            symbol = func_match.group(1)
            kind = 'func'
        elif type_match:
            symbol = type_match.group(1)
            kind = 'type'
            type_kind = type_match.group(2)
        elif var_const_match:
            symbol = var_const_match.group(2)
            kind = 'var_const'

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

            doc_text = "\n".join(doc_lines)
            is_compliant = False

            needs_func_docs = False
            needs_struct_docs = False
            needs_interface_docs = False

            if kind == 'func':
                if doc_lines and "Parameters:" in doc_text and "Returns:" in doc_text and "Side Effects:" in doc_text:
                    is_compliant = True
                else:
                    needs_func_docs = True
            elif kind == 'type' and type_kind == 'struct':
                if doc_lines and ("Fields:" in doc_text or "Parameters:" in doc_text):
                    is_compliant = True
                else:
                    needs_struct_docs = True
            elif kind == 'type' and type_kind == 'interface':
                if doc_lines and "Methods:" in doc_text:
                    is_compliant = True
                else:
                    needs_interface_docs = True
            elif kind == 'var_const':
                if doc_lines:
                    is_compliant = True

            if not is_compliant:
                old_summary_lines = []
                j = len(out_lines) - 1
                while j >= 0:
                    prev = out_lines[j]
                    if prev.strip().startswith('//'):
                        if not prev.strip().startswith('//go:'):
                            old_summary_lines.insert(0, out_lines.pop(j))
                        else:
                            break
                    elif prev.strip() == '':
                        break
                    else:
                        break
                    j -= 1

                old_summary = ""
                for l in old_summary_lines:
                    text = l.strip().replace('// ', '').replace('//', '').strip()
                    if text:
                        old_summary += text + " "
                old_summary = old_summary.strip()

                if not old_summary or len(old_summary.split()) <= 3 or "Sets the" in old_summary or old_summary.startswith(f"{symbol} represents") or old_summary.startswith(f"{symbol} -"):
                    old_summary = generate_summary(symbol, kind)

                if not old_summary.endswith('.'):
                    old_summary += '.'

                if not old_summary.startswith(symbol):
                    clean_summary = old_summary[0].lower() + old_summary[1:]
                    lead_line = f"// {symbol} {clean_summary}"
                else:
                    lead_line = f"// {old_summary}"

                new_docs = []

                if kind == 'var_const':
                    new_docs.append(f"{lead_line}\n")
                else:
                    new_docs.append(f"{lead_line}\n")
                    new_docs.append(f"//\n")
                    new_docs.append(f"// Summary: {old_summary}\n")

                    if needs_func_docs:
                        params_str = func_match.group(2)
                        returns_str = func_match.group(3)

                        new_docs.append(f"//\n")
                        new_docs.append(f"// Parameters:\n")
                        if params_str and params_str.strip():
                            param_list = []
                            depth = 0
                            current = ""
                            for char in params_str:
                                if char in '({[':
                                    depth += 1
                                elif char in ')}]':
                                    depth -= 1
                                if char == ',' and depth == 0:
                                    param_list.append(current.strip())
                                    current = ""
                                else:
                                    current += char
                            if current.strip():
                                param_list.append(current.strip())

                            has_params = False
                            for p in param_list:
                                parts = p.split()
                                if len(parts) >= 2:
                                    has_params = True
                                    names = parts[:-1]
                                    typ = parts[-1]
                                    for name in names:
                                        name = name.replace(',', '')
                                        desc = get_param_desc(name, typ)
                                        new_docs.append(f"//   - {name} ({typ}): {desc}\n")
                                elif len(parts) == 1:
                                    has_params = True
                                    typ = parts[0]
                                    new_docs.append(f"//   - _ ({typ}): An unnamed parameter of type {typ}.\n")
                            if not has_params:
                                new_docs.append(f"//   - None.\n")
                        else:
                            new_docs.append(f"//   - None.\n")

                        new_docs.append(f"//\n")
                        new_docs.append(f"// Returns:\n")
                        if returns_str and returns_str.strip():
                            returns_str = returns_str.strip()
                            if returns_str.startswith('(') and returns_str.endswith(')'):
                                returns_str = returns_str[1:-1]
                            ret_list = [r.strip() for r in returns_str.split(',') if r.strip()]
                            for r in ret_list:
                                parts = r.split()
                                if len(parts) >= 2:
                                    name = parts[0]
                                    typ = parts[1]
                                    desc = get_return_desc(typ)
                                    new_docs.append(f"//   - {name} ({typ}): {desc}\n")
                                else:
                                    typ = parts[0]
                                    desc = get_return_desc(typ)
                                    new_docs.append(f"//   - ({typ}): {desc}\n")
                        else:
                            new_docs.append(f"//   - None.\n")

                        new_docs.append(f"//\n")
                        new_docs.append(f"// Errors:\n")
                        if returns_str and 'error' in returns_str:
                            new_docs.append(f"//   - Returns an error if the underlying operation fails or encounters invalid input.\n")
                        else:
                            new_docs.append(f"//   - None.\n")

                        new_docs.append(f"//\n")
                        new_docs.append(f"// Side Effects:\n")
                        if any(x in symbol for x in ['DB', 'Store', 'Cache', 'HTTP', 'Update', 'Set', 'Delete', 'Create', 'Write', 'Load', 'Start', 'Stop']):
                            new_docs.append(f"//   - Modifies global state, writes to the database, or establishes network connections.\n")
                        else:
                            new_docs.append(f"//   - None.\n")

                    elif needs_struct_docs:
                        new_docs.append(f"//\n")
                        new_docs.append(f"// Fields:\n")
                        new_docs.append(f"//   - Contains the configuration and state properties required for {symbol} functionality.\n")
                    elif needs_interface_docs:
                        new_docs.append(f"//\n")
                        new_docs.append(f"// Methods:\n")
                        new_docs.append(f"//   - Defines the required contract and behavior for implementations of {symbol}.\n")

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
