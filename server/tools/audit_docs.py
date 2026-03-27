#!/usr/bin/env python3
import os, re, sys

FUNC_PATTERN = re.compile(r'^func\s+([A-Z]\w*)\s*\((.*?)\)\s*(.*)\{')
METHOD_PATTERN = re.compile(r'^func\s+\([^)]+\)\s+([A-Z]\w*)\s*\((.*?)\)\s*(.*)\{')
TYPE_PATTERN = re.compile(r'^type\s+([A-Z]\w*)\s+')
VAR_CONST_PATTERN = re.compile(r'^(var|const)\s+([A-Z]\w*)\s+')

def has_section(doc_lines, section_name):
    for line in doc_lines:
        if f"{section_name}:" in line: return True
    return False

def check_file(filepath):
    with open(filepath, 'r') as f: lines = f.readlines()
    violations = []
    for i, line in enumerate(lines):
        line = line.rstrip()
        symbol = None
        kind = None
        returns_str = ""
        m_func = FUNC_PATTERN.match(line)
        m_method = METHOD_PATTERN.match(line)
        m_type = TYPE_PATTERN.match(line)
        m_var_const = VAR_CONST_PATTERN.match(line)
        if m_func:
            symbol, kind, returns_str = m_func.group(1), "func", m_func.group(3)
        elif m_method:
            symbol, kind, returns_str = m_method.group(1), "method", m_method.group(3)
        elif m_type:
            symbol, kind = m_type.group(1), "type"
        elif m_var_const:
            symbol, kind = m_var_const.group(2), "var/const"
        if symbol:
            doc_lines = []
            j = i - 1
            while j >= 0:
                prev = lines[j].strip()
                if prev.startswith('//'):
                    if not (prev.startswith('//go:') or prev.startswith('// +build')):
                        doc_lines.insert(0, prev)
                elif prev == '': pass
                else: break
                j -= 1
            reasons = []
            if not doc_lines: reasons.append("Missing documentation")
            else:
                if not has_section(doc_lines, "Summary"): reasons.append("Missing 'Summary:'")
                if kind in ("func", "method"):
                    if not has_section(doc_lines, "Parameters"): reasons.append("Missing 'Parameters:'")
                    if (returns_str.strip() and returns_str.strip() != "{") and not has_section(doc_lines, "Returns"):
                        reasons.append("Missing 'Returns:'")
                    if not has_section(doc_lines, "Errors"): reasons.append("Missing 'Errors:'")
                    if not has_section(doc_lines, "Side Effects"): reasons.append("Missing 'Side Effects:'")
            if reasons: violations.append((i + 1, symbol, ", ".join(reasons)))
    return violations

def scan_dir(root_dir):
    count = 0
    for root, _, files in os.walk(root_dir):
        if any(x in root for x in ["vendor", "node_modules"]): continue
        for file in files:
            if file.endswith('.go') and not file.endswith('_test.go'):
                path = os.path.join(root, file)
                v = check_file(path)
                if v:
                    print(f"File: {path}")
                    for line, sym, reason in v:
                        print(f"  Line {line}: {sym} ({reason})")
                        count += 1
    return count

if __name__ == '__main__':
    dirs = sys.argv[1:] if len(sys.argv) > 1 else ['server/pkg', 'server/cmd']
    total = sum(scan_dir(d) for d in dirs)
    if total > 0:
        print(f"\nFound {total} violations.")
        sys.exit(1)
    else:
        print("\nAll public interfaces are documented correctly!")
        sys.exit(0)
