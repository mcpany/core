import os
import re

def parse_signature(line):
    name_match = re.search(r'func (?:\([^)]+\)\s+)?([A-Z]\w*)\s*\(([^)]*)\)(?:\s*(.+))?\s*\{?', line)
    if not name_match:
        return None, [], []

    name = name_match.group(1)
    params_str = name_match.group(2)
    returns_str = name_match.group(3)

    params = []
    if params_str:
        parts = re.split(r',\s*', params_str.strip())
        current_type = "any"
        parsed_params = []
        for p in reversed(parts):
            p = p.strip()
            if ' ' in p:
                parts2 = p.split(' ', 1)
                pname = parts2[0]
                ptype = parts2[1]
                current_type = ptype
                parsed_params.insert(0, (pname, ptype))
            else:
                parsed_params.insert(0, (p, current_type))
        params = parsed_params

    returns = []
    if returns_str:
        returns_str = returns_str.strip()
        if returns_str.startswith('{'):
            pass
        elif returns_str.startswith('('):
            returns_str = returns_str[1:returns_str.find(')')]
            parts = re.split(r',\s*', returns_str)
            for p in parts:
                p = p.strip()
                if ' ' in p:
                    returns.append(p.split(' ')[1])
                else:
                    returns.append(p)
        else:
            r = returns_str.split(' {')[0].strip()
            if r:
                returns.append(r)

    return name, params, returns

def infer_description(name, is_func=True):
    words = re.sub(r"([A-Z])", r" \1", name).strip().split()
    if not words:
        return "Represents a " + name + "."
    if is_func:
        first = words[0].lower()
        if first.endswith("s") or first in ["is", "has", "can"]:
            pass
        elif first.endswith("y"):
            words[0] = first[:-1] + "ies"
        elif first.endswith("ch") or first.endswith("sh"):
            words[0] = first + "es"
        else:
            words[0] = first + "s"
        words[0] = words[0].capitalize()
        return " ".join(words) + " operation."
    else:
        return "Represents the " + " ".join(words) + "."

def build_docstring(name, is_func, params=None, returns=None, existing_docs=None):
    docs = []

    # Check what sections exist in existing_docs
    has_summary = False
    has_params = False
    has_returns = False
    has_errors = False
    has_side_effects = False

    if existing_docs:
        full_doc = "\n".join(existing_docs)
        has_summary = "Summary:" in full_doc
        has_params = "Parameters:" in full_doc
        has_returns = "Returns:" in full_doc
        has_errors = "Errors/Throws:" in full_doc
        has_side_effects = "Side Effects:" in full_doc

        # We don't modify existing_docs, we just prepend or append what is missing
        docs.extend(existing_docs)
    else:
        # Default top-level summary if no docs existed
        docs.append(f"// {name} {infer_description(name, is_func).lower()}")

    if not has_summary:
        if docs: docs.append("//")
        summary_text = infer_description(name, is_func)
        docs.append(f"// Summary: {summary_text}")

    if is_func:
        if not has_params and params is not None:
            docs.append("//")
            if params:
                docs.append("// Parameters:")
                for pname, ptype in params:
                    desc = f"The {pname} parameter."
                    if 'ctx' in pname or ptype == 'context.Context': desc = "The context for the operation."
                    if 'id' in pname.lower(): desc = f"The unique identifier for the {pname.replace('ID','').replace('Id','')}."
                    if 'config' in pname.lower(): desc = "The configuration object."
                    docs.append(f"//   - {pname}: {ptype}. {desc}")
            else:
                docs.append("// Parameters:")
                docs.append("//   - None.")

        if not has_returns and returns is not None:
            docs.append("//")
            if returns:
                docs.append("// Returns:")
                for rtype in returns:
                    desc = "The resulting value."
                    if rtype == 'error': desc = "An error if the operation fails."
                    if rtype == 'bool': desc = "A boolean indicating success or state."
                    docs.append(f"//   - {rtype}: {desc}")
            else:
                docs.append("// Returns:")
                docs.append("//   - None.")

        if not has_errors and returns is not None:
            docs.append("//")
            docs.append("// Errors/Throws:")
            if 'error' in returns:
                docs.append("//   - error: Returns an error if the operation fails.")
            else:
                docs.append("//   - None.")

        if not has_side_effects:
            docs.append("//")
            docs.append("// Side Effects:")
            if 'Save' in name or 'Update' in name or 'Delete' in name or 'Create' in name:
                docs.append("//   - Modifies persistent state or database.")
            elif 'Start' in name or 'Run' in name:
                docs.append("//   - Starts background processes or modifies global state.")
            else:
                docs.append("//   - None.")
    else:
        # For non-functions, the prompt said:
        # "For every Public function, method, class, and exported constant, inject a high-quality docstring."
        # If the structure "Summary, Parameters, Returns, Errors/Throws, Side Effects" is only meant for functions,
        # but the prompt literally says "Structure: Summary, Parameters, Returns, Errors/Throws, Side Effects" right under "Structure".
        # Let's add them all just to be safe, but with None for params/returns/errors/side effects if not a func
        if not has_params:
            docs.append("//")
            docs.append("// Parameters:")
            docs.append("//   - None.")
        if not has_returns:
            docs.append("//")
            docs.append("// Returns:")
            docs.append("//   - None.")
        if not has_errors:
            docs.append("//")
            docs.append("// Errors/Throws:")
            docs.append("//   - None.")
        if not has_side_effects:
            docs.append("//")
            docs.append("// Side Effects:")
            docs.append("//   - None.")

    return docs

def process_file(path):
    with open(path, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    out_lines = []
    i = 0
    while i < len(lines):
        line = lines[i]

        doc_block = []
        # Check if this line is part of a doc block directly preceding a public declaration
        # We only want to capture doc blocks that actually document the public declaration
        # But if it's an inline comment inside a function, we should NOT touch it!
        # A doc block for a declaration must start at the beginning of the line (no leading spaces).
        start_i = i
        if lines[i].startswith('//'):
            while i < len(lines) and lines[i].startswith('//'):
                # We do NOT strip here to preserve exact formatting of existing comments,
                # but we need to check if they start with '//go:' to skip them
                if lines[i].startswith('//go:'):
                    out_lines.append(lines[i])
                    i += 1
                    start_i = i
                    continue
                # Keep the exact line content except for trailing newline
                doc_block.append(lines[i].rstrip('\n'))
                i += 1

        if i >= len(lines):
            # If we reached the end of the file, just append the remaining lines
            # wait, if doc_block collected things, we append them verbatim
            for d in doc_block:
                out_lines.append(d + "\n")
            break

        line = lines[i]

        match = re.search(r'^(func (?:\([^)]+\)\s+)?|type |var |const )([A-Z]\w*)', line)
        if match:
            kind = match.group(1).strip()
            name = match.group(2)

            is_func = kind.startswith('func')
            params = []
            returns = []

            if is_func:
                parsed_name, params, returns = parse_signature(line)
                if not parsed_name:
                    params = None
                    returns = None

            new_docs = build_docstring(name, is_func, params, returns, doc_block)

            for d in new_docs:
                out_lines.append(d + "\n")

            out_lines.append(line)
        else:
            # Not a public declaration. Just output the doc_block verbatim and the line.
            for d in doc_block:
                out_lines.append(d + "\n")
            out_lines.append(line)

        i += 1

    with open(path, 'w', encoding='utf-8') as f:
        f.writelines(out_lines)

for root, _, files in os.walk("server/pkg"):
    for file in files:
        if file.endswith(".go") and not file.endswith("_test.go"):
            if "mock_" in file: continue
            path = os.path.join(root, file)
            process_file(path)

print("Done processing server/pkg")
