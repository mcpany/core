# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import sys
import re
import os

def parse_params(params_str):
    """
    Tries to parse parameters from a Go function signature string.
    Returns a list of (name, type) tuples.
    This is a heuristic parser.
    """
    params = []
    if not params_str.strip():
        return params

    parts = params_str.split(',')

    current_names = []
    for part in parts:
        part = part.strip()
        if not part: continue

        if ' ' in part:
            subparts = part.rsplit(' ', 1)
            if len(subparts) == 2:
                name = subparts[0].strip()
                typ = subparts[1].strip()

                for n in current_names:
                    params.append((n, typ))
                current_names = []

                params.append((name, typ))
            else:
                 pass
        else:
            current_names.append(part)

    return params

def find_funcs(content):
    matches = []

    # Regex to find 'func (receiver) Name' or 'func Name'
    # We want to capture the name
    # Must be at start of line (allowing for tabs/spaces)
    start_regex = re.compile(r'^\s*func\s+(?:\(.*\)\s+)?([A-Z][a-zA-Z0-9_]*)', re.MULTILINE)

    last_idx = 0
    while True:
        m = start_regex.search(content, last_idx)
        if not m:
            break

        func_name = m.group(1)
        func_start = m.start()
        last_idx = m.end()

        # Scan forward to find params
        idx = m.end()
        while idx < len(content) and content[idx].isspace():
            idx += 1

        if idx >= len(content) or content[idx] != '(':
            continue

        # Parse params (balanced parens)
        params_start = idx + 1
        idx += 1
        depth = 1
        while idx < len(content) and depth > 0:
            if content[idx] == '(':
                depth += 1
            elif content[idx] == ')':
                depth -= 1
            idx += 1

        if depth != 0:
             continue # Unbalanced

        params_end = idx - 1
        params_str = content[params_start:params_end]

        # Parse returns (until {)
        returns_start = idx
        while idx < len(content) and content[idx] != '{':
            idx += 1

        if idx >= len(content):
             continue # No body? Interface?

        returns_str = content[returns_start:idx].strip()

        # Comment finding
        # Look backwards from func_start
        # We need to be careful with indices.
        # Check lines before func_start

        prefix = content[:func_start]
        lines = prefix.splitlines(keepends=True)
        comment_lines = []

        # Iterate backwards
        # If the last line of prefix doesn't end with newline, it means func starts on same line (weird but possible)
        # But splitlines keeps ends.

        # Check if there is whitespace between last line end and func_start?
        # func_start is index in content.
        # sum(len(l) for l in lines) == func_start (roughly, if no trailing chars on last line)

        # Simpler: traverse backwards from func_start-1

        c_idx = func_start - 1
        while c_idx >= 0 and content[c_idx].isspace():
            c_idx -= 1

        # Now c_idx is at non-space or end of comment
        # We need to extract the comment block.
        # This is getting tricky to match exact Go rules.
        # Let's use the line-based approach which is safer.

        found_comment = ""
        comment_end_pos = func_start

        # Re-read file lines?
        # No, just look at the lines list.
        # If the last line contains only whitespace, skip it?

        if lines:
             # Check if last line is empty/whitespace
             if not lines[-1].strip():
                  lines.pop()

        for line in reversed(lines):
            stripped = line.strip()
            if stripped.startswith('//'):
                comment_lines.insert(0, line)
            elif not stripped:
                 break
            else:
                 break

        comment_block = "".join(comment_lines)

        # Calculate insert position
        # If comment exists, insert at end of comment block.
        # If no comment, insert at func_start.

        # We need the exact index of where the comment block ends in 'content' to append to it.
        # Or where it starts to replace it (not doing replace).

        # For appending: we need the index right after the last char of comment_block.
        # But wait, we extracted comment_block from lines.
        # Finding exact index is hard if we just use lines.

        # Let's just store the comment_block string and use it for logic.
        # For insertion index:
        # If comment exists, we want to insert BEFORE the newline that terminates the comment block?
        # No, we want to append TO the comment block.
        # Ideally we insert at `func_start` but prefixed with `\n`? No.

        # Let's stick to:
        # If comment exists, we use regex to find it again near func_start?
        # Or just use the fact that comment immediately precedes func (ignoring whitespace).

        matches.append({
            'func_name': func_name,
            'params_str': params_str,
            'returns_str': returns_str,
            'comment_block': comment_block,
            'func_start': func_start
        })

    return matches

def fix_go_file(filepath, apply=False):
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
    except Exception as e:
        print(f"Error reading {filepath}: {e}")
        return

    funcs = find_funcs(content)
    modifications = []

    for f in funcs:
        comment_block = f['comment_block']
        func_name = f['func_name']
        params_str = f['params_str']
        returns_str = f['returns_str']
        func_start = f['func_start']

        needs_summary = False
        needs_params = False
        needs_returns = False
        needs_errors = False

        existing_comment = comment_block if comment_block else ""

        if not existing_comment:
            needs_summary = True

        has_params_section = "Parameters:" in existing_comment
        has_returns_section = "Returns:" in existing_comment
        has_errors_section = "Errors:" in existing_comment or "Throws/Errors:" in existing_comment

        if params_str.strip():
             needs_params = not has_params_section

        if returns_str and returns_str.strip():
             needs_returns = not has_returns_section
             if "error" in returns_str:
                  needs_errors = not has_errors_section

        if not (needs_summary or needs_params or needs_returns or needs_errors):
             continue

        new_doc = []

        if needs_summary:
             new_doc.append(f"// {func_name} ...")
             new_doc.append("//")

        if needs_params:
             new_doc.append("// Parameters:")
             params = parse_params(params_str)
             if params:
                 for p_name, p_type in params:
                     new_doc.append(f"//   - {p_name}: {p_type}.")
             else:
                 new_doc.append("//   - (none)")
             new_doc.append("//")

        if needs_returns:
             new_doc.append("// Returns:")
             new_doc.append(f"//   - (results)")
             new_doc.append("//")

        if needs_errors:
             new_doc.append("// Errors:")
             new_doc.append("//   - error: An error if the operation fails.")
             new_doc.append("//")

        if not new_doc:
             continue

        if new_doc[-1] == "//":
             new_doc.pop()

        insertion = "\n".join(new_doc) + "\n"

        if not comment_block:
             modifications.append((func_start, insertion))
        else:
             # Find where to append.
             # We want to insert after the comment block.
             # Since comment_block was extracted from text before func_start,
             # its end is effectively near func_start (minus whitespace).

             # But wait, if we append *after* comment block, there might be a blank line between comment and func?
             # Go doc requires comment to be immediately adjacent.

             # So we must append *to* the comment block.
             # Finding the end of the comment block in `content`:
             # It is the index `func_start` minus any whitespace lines.

             # Actually, simpler: Insert at `func_start`, but start with `//`?
             # No, if we insert at `func_start`, it pushes `func` down.
             # The comment block is *above*.
             # So we insert at `func_start`.
             # BUT, if we insert there, it becomes separated from the previous comment?
             # existing:
             # // comment
             # func Name

             # we insert:
             # // comment
             # // Parameters: ...
             # func Name

             # This works perfectly!

             # But we need to make sure we don't insert a blank line if `insertion` starts with `//`.
             # `insertion` = "// Parameters:...\n"

             # If existing comment ends with `\n`, and we insert at `func_start`,
             # we get `// comment\n// Parameters...`
             # This is correct.

             modifications.append((func_start, insertion))

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
        print("Usage: python tools/doc_fixer.py <file_or_dir> [--apply]")
        sys.exit(1)

    target = sys.argv[1]
    apply = "--apply" in sys.argv

    if os.path.isfile(target):
        fix_go_file(target, apply)
    elif os.path.isdir(target):
        for root, _, files in os.walk(target):
            for file in files:
                if file.endswith(".go") and not file.endswith("_test.go") and not file.endswith(".pb.go"):
                    fix_go_file(os.path.join(root, file), apply)
