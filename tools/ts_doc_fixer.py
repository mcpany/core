# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import sys
import re
import os

def parse_ts_params(params_str):
    # Extremely basic parser
    # "a: string, b: number" -> [("a", "string"), ("b", "number")]
    # "props: Props" -> [("props", "Props")]
    # "{ a, b }: Props" -> [("root0", "Props")] (destructuring is hard)

    params = []
    if not params_str.strip():
        return params

    # Split by comma (ignoring braces)
    # Reuse Go logic somewhat?

    # Just split by comma for now.
    parts = params_str.split(',')
    for p in parts:
        p = p.strip()
        if not p: continue
        if ':' in p:
            name, typ = p.split(':', 1)
            params.append((name.strip(), typ.strip()))
        else:
            params.append((p, "any"))

    return params

def find_ts_exports(content):
    matches = []

    # Regex to find 'export ...'
    start_regex = re.compile(r'export\s+(?:default\s+)?(function|const|class|type|interface)\s+([a-zA-Z0-9_]+)')

    last_idx = 0
    while True:
        m = start_regex.search(content, last_idx)
        if not m:
            break

        kind = m.group(1)
        name = m.group(2)
        start_pos = m.start()
        last_idx = m.end()

        params_str = ""
        returns_str = "" # Not easily parseable in TS without full parser

        # If function or const (arrow func), try to find params
        if kind == 'function':
            # Scan for (
            idx = m.end()
            while idx < len(content) and content[idx].isspace():
                idx += 1
            if idx < len(content) and content[idx] == '(':
                # Find matching )
                p_start = idx + 1
                idx += 1
                depth = 1
                while idx < len(content) and depth > 0:
                    if content[idx] == '(': depth += 1
                    elif content[idx] == ')': depth -= 1
                    idx += 1
                params_str = content[p_start:idx-1]

        elif kind == 'const':
            # export const Name = (...) => ...
            # or export const Name = function(...) ...
            # Find =
            idx = m.end()
            while idx < len(content) and content[idx] != '=':
                idx += 1
            idx += 1 # skip =
            while idx < len(content) and content[idx].isspace():
                idx += 1

            if idx < len(content):
                if content[idx] == '(':
                     # Arrow func params
                    p_start = idx + 1
                    idx += 1
                    depth = 1
                    while idx < len(content) and depth > 0:
                        if content[idx] == '(': depth += 1
                        elif content[idx] == ')': depth -= 1
                        idx += 1
                    params_str = content[p_start:idx-1]
                elif content[idx:idx+8] == 'function':
                     # function params
                     # skip function
                     idx += 8
                     while idx < len(content) and content[idx].isspace(): idx += 1
                     if idx < len(content) and content[idx] == '(':
                        p_start = idx + 1
                        idx += 1
                        depth = 1
                        while idx < len(content) and depth > 0:
                            if content[idx] == '(': depth += 1
                            elif content[idx] == ')': depth -= 1
                            idx += 1
                        params_str = content[p_start:idx-1]

        # Find preceding comment
        # Scan backwards from start_pos

        # Similar logic to Go fixer
        prefix = content[:start_pos]
        lines = prefix.splitlines(keepends=True)
        if lines and not lines[-1].strip(): lines.pop()

        comment_lines = []
        for line in reversed(lines):
            stripped = line.strip()
            if stripped.startswith('//') or stripped.startswith('*') or stripped.startswith('/*'):
                comment_lines.insert(0, line)
            elif stripped == '*/': # End of multiline
                 comment_lines.insert(0, line)
                 # Need to find /*
                 # This simple logic fails for multiline block end.
                 # Let's just look for `/**` block end?
                 pass
            elif not stripped:
                 break
            else:
                 break

        # Better comment extraction for JSDoc
        # Look for `*/` before `start_pos`.
        c_end = start_pos - 1
        while c_end >= 0 and content[c_end].isspace():
            c_end -= 1

        comment_block = ""
        if c_end >= 1 and content[c_end] == '/' and content[c_end-1] == '*':
            # Found end of block comment */
            # Search backwards for /**
            c_start = content.rfind('/**', 0, c_end)
            if c_start != -1:
                comment_block = content[c_start:c_end+1]

        matches.append({
            'name': name,
            'kind': kind,
            'params_str': params_str,
            'comment_block': comment_block,
            'start_pos': start_pos
        })

    return matches

def fix_ts_file(filepath, apply=False):
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
    except Exception as e:
        print(f"Error reading {filepath}: {e}")
        return

    exports = find_ts_exports(content)
    modifications = []

    for ex in exports:
        name = ex['name']
        kind = ex['kind']
        params_str = ex['params_str']
        comment_block = ex['comment_block']
        start_pos = ex['start_pos']

        if kind not in ['function', 'const']:
            continue # Skip classes/interfaces for now unless easy

        # Analyze existing comment
        needs_doc = False
        if not comment_block or "/**" not in comment_block:
            needs_doc = True

        has_params_doc = "@param" in comment_block if comment_block else False
        has_returns_doc = "@returns" in comment_block if comment_block else False

        needs_params = False
        if params_str.strip():
             needs_params = not has_params_doc

        # Always add returns for functions?
        needs_returns = not has_returns_doc

        if not (needs_doc or needs_params or needs_returns):
            continue

        new_doc = []
        if needs_doc:
            new_doc.append(f"/**")
            new_doc.append(f" * {name} ...")
            new_doc.append(f" *")

        if needs_params:
            params = parse_ts_params(params_str)
            for p_name, p_type in params:
                new_doc.append(f" * @param {p_name} - {p_type}.")

        if needs_returns:
            new_doc.append(f" * @returns ...")

        if needs_doc:
            new_doc.append(f" */")

        if not new_doc:
            continue

        insertion = "\n".join(new_doc) + "\n"

        if not comment_block:
             modifications.append((start_pos, insertion))
        else:
             # Appending to JSDoc is harder because we need to insert before `*/`.
             # And ensure formatting.
             # For now, only handle missing doc cases to avoid corruption.
             pass

    if not modifications:
        return

    modifications.sort(key=lambda x: x[0], reverse=True)

    new_content = content
    for idx, text in modifications:
        new_content = new_content[:idx] + text + new_content[idx:]

    if apply:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(new_content)
        print(f"Fixed {filepath}")
    else:
        print(f"Would fix {filepath} ({len(modifications)} insertions)")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python tools/ts_doc_fixer.py <file_or_dir> [--apply]")
        sys.exit(1)

    target = sys.argv[1]
    apply = "--apply" in sys.argv

    if os.path.isfile(target):
        fix_ts_file(target, apply)
    elif os.path.isdir(target):
        for root, _, files in os.walk(target):
            for file in files:
                if (file.endswith(".ts") or file.endswith(".tsx")) and not file.endswith(".d.ts") and not file.endswith(".test.ts"):
                    fix_ts_file(os.path.join(root, file), apply)
