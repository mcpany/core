import os
import re

def is_exported(name):
    return name and name[0].isupper()

missing = []

for root, _, files in os.walk('./server'):
    if 'test' in root or 'vendor' in root:
        continue
    for file in files:
        if file.endswith('.go') and not 'test' in file and not file.startswith('.'):
            path = os.path.join(root, file)
            with open(path, 'r', encoding='utf-8') as f:
                content = f.read()
                lines = content.split('\n')
                for i, line in enumerate(lines):
                    # check for exported func, type, struct, interface
                    match = re.match(r'^func\s+(?:\([^)]+\)\s+)?([A-Z]\w*)\(', line)
                    if not match:
                        match = re.match(r'^type\s+([A-Z]\w*)\s+(struct|interface|string|int|bool)', line)
                    if match:
                        name = match.group(1)
                        if is_exported(name):
                            # check previous lines for doc
                            has_doc = False
                            for j in range(1, 4):
                                if i-j >= 0 and lines[i-j].startswith('//'):
                                    has_doc = True
                                    break
                            if not has_doc:
                                missing.append(f"{path}:{i+1} {name}")
print(len(missing))
for m in missing[:20]:
    print(m)
