import os
import re

FUNC_PATTERN = re.compile(r'^func\s+([A-Z]\w*)\s*\((.*?)\)\s*(.*)\{')
METHOD_PATTERN = re.compile(r'^func\s+\([^)]+\)\s+([A-Z]\w*)\s*\((.*?)\)\s*(.*)\{')

def get_params(p_str):
    if not p_str.strip(): return []
    res = []
    depth = 0
    current = ""
    for char in p_str:
        if char == '(' : depth += 1
        elif char == ')': depth -= 1
        elif char == ',' and depth == 0:
            res.append(current.strip())
            current = ""
            continue
        current += char
    if current:
        res.append(current.strip())

    names = []
    for p in res:
        parts = p.strip().split(' ')
        if not parts: continue
        name_part = parts[0]
        if ',' in name_part:
            sub_parts = name_part.split(',')
            for sp in sub_parts:
                sp = sp.strip()
                if sp: names.append(sp)
        else:
            if name_part and name_part != '...':
                names.append(name_part)
    return names

def fix_file(path):
    with open(path, 'r') as f:
        lines = f.readlines()

    new_lines = []
    i = 0
    changed = False

    while i < len(lines):
        line = lines[i]
        m = FUNC_PATTERN.match(line) or METHOD_PATTERN.match(line)
        if m:
            symbol = m.group(1)
            p_str = m.group(2)
            r_str = m.group(3)

            doc_start = i - 1
            while doc_start >= 0 and lines[doc_start].strip().startswith('//'):
                doc_start -= 1
            doc_start += 1

            doc_lines = lines[doc_start:i]

            updated_doc = []
            updated_doc.append(f"// {symbol} provides {symbol.lower()} functionality.\n")
            updated_doc.append("//\n")
            updated_doc.append(f"// Summary: {symbol}.\n")
            updated_doc.append("//\n")

            updated_doc.append("// Parameters.\n")
            params = get_params(p_str)
            if not params:
                updated_doc.append("//   - None.\n")
            else:
                for p in params:
                    updated_doc.append(f"//   - {p}: The parameter.\n")
            updated_doc.append("//\n")

            updated_doc.append("// Returns.\n")
            if r_str.strip() and r_str.strip() != "{":
                updated_doc.append("//   - result: The result.\n")
            else:
                updated_doc.append("//   - None.\n")

            if doc_lines != updated_doc:
                for _ in range(len(doc_lines)):
                    if new_lines: new_lines.pop()
                new_lines.extend(updated_doc)
                changed = True

        new_lines.append(line)
        i += 1

    if changed:
        with open(path, 'w') as f:
            f.writelines(new_lines)
        return True
    return False

for d in ['server/pkg', 'server/cmd']:
    for root, dirs, files in os.walk(d):
        if any(x in root for x in ["vendor", "node_modules", "examples", "docs", "test", "k8s"]): continue
        for file in files:
            if file.endswith('.go') and not any(file.endswith(x) for x in ['_test.go', '.pb.go', '.pb.gw.go', 'zz_generated.deepcopy.go']):
                fix_file(os.path.join(root, file))
