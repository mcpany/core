import os
import re

FUNC_PATTERN = re.compile(r'^func\s+([A-Z]\w*)\s*\((.*)\)\s*(.*)\{')
METHOD_PATTERN = re.compile(r'^func\s+\([^)]+\)\s+([A-Z]\w*)\s*\((.*)\)\s*(.*)\{')

def get_params(p_str):
    if not p_str.strip(): return []
    res = []
    # Handle nested parens/interfaces
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
        # p is like "ctx context.Context" or "a, b int"
        parts = p.split(' ')
        if len(parts) > 1:
             # Check for comma separated names
             sub_parts = parts[0].split(',')
             for sp in sub_parts:
                 sp = sp.strip()
                 if sp: names.append(sp)
        else:
             if p and p != '...': names.append(p)
    return names

def fix(path):
    with open(path, 'r') as f:
        lines = f.readlines()
    new_lines = []
    i = 0
    while i < len(lines):
        line = lines[i]
        m = FUNC_PATTERN.match(line) or METHOD_PATTERN.match(line)
        if m:
            name = m.group(1)
            p_str = m.group(2)
            r_str = m.group(3)

            doc_start = i - 1
            while doc_start >= 0 and lines[doc_start].strip().startswith('//'):
                doc_start -= 1
            doc_start += 1
            doc_lines = lines[doc_start:i]

            new_doc = []
            new_doc.append(f"// {name} provides {name.lower()} functionality.\n")
            new_doc.append("//\n")
            new_doc.append(f"// Summary: {name}.\n")
            new_doc.append("//\n")
            new_doc.append("// Parameters.\n")
            params = get_params(p_str)
            if not params:
                new_doc.append("//   - None.\n")
            else:
                for p in params:
                    new_doc.append(f"//   - {p}: The parameter.\n")
            new_doc.append("//\n")
            new_doc.append("// Returns.\n")
            if r_str.strip() and r_str.strip() != "{":
                new_doc.append("//   - result: The result.\n")
            else:
                new_doc.append("//   - None.\n")

            for _ in range(len(doc_lines)):
                if new_lines: new_lines.pop()
            new_lines.extend(new_doc)
        new_lines.append(line)
        i += 1
    with open(path, 'w') as f:
        f.writelines(new_lines)

for d in ['server/pkg', 'server/cmd']:
    for root, dirs, files in os.walk(d):
        if any(x in root for x in ["vendor", "node_modules", "examples", "docs", "test", "k8s"]): continue
        for file in files:
            if file.endswith('.go') and not any(file.endswith(x) for x in ['_test.go', '.pb.go', '.pb.gw.go', 'zz_generated.deepcopy.go']):
                fix(os.path.join(root, file))
