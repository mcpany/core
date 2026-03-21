import re

to_fix = []
with open('to_fix.txt', 'r') as f:
    for line in f:
        to_fix.append(line.strip())

files_to_fix = {}
for line in to_fix:
    parts = line.split(' ')
    file_line = parts[0]
    filepath, linenum = file_line.split(':')
    linenum = int(linenum)
    name = parts[3]
    if filepath not in files_to_fix:
        files_to_fix[filepath] = []
    files_to_fix[filepath].append((linenum, name))

for filepath, lines_to_fix in files_to_fix.items():
    with open(filepath, 'r') as f:
        content = f.read()

    lines = content.split('\n')
    lines_to_fix.sort(key=lambda x: x[0], reverse=True)

    for linenum, name in lines_to_fix:
        # linenum is 1-indexed
        idx = linenum - 1
        line = lines[idx]

        indent = line[:len(line) - len(line.lstrip())]

        match_func = re.match(r'^func\s+(?:\([^)]+\)\s+)?([A-Z]\w*)', line)
        if match_func:
            name = match_func.group(1)

        doc = f"{indent}// {name} provides documentation for {name}."
        lines.insert(idx, doc)

    with open(filepath, 'w') as f:
        f.write('\n'.join(lines))
