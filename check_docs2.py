import os
import re

def check_file(path):
    with open(path, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    in_doc = False
    doc_block = []

    for i, line in enumerate(lines):
        if line.strip().startswith('//'):
            in_doc = True
            doc_block.append(line.strip())
        else:
            if line.startswith('func ') or line.startswith('type ') or line.startswith('var ') or line.startswith('const '):
                match = re.search(r'(func (?:\([^)]+\)\s+)?|type |var |const )([A-Z]\w*)', line)
                if match:
                    name = match.group(2)
                    if not in_doc:
                        print(f"{path}:{i+1} Missing entirely doc for {name}")
                    else:
                        doc_text = '\n'.join(doc_block)
                        has_summary = 'Summary:' in doc_text

                        if not has_summary:
                            print(f"{path}:{i+1} Missing 'Summary:' for {name}")
            in_doc = False
            doc_block = []

for root, _, files in os.walk("server"):
    for file in files:
        if file.endswith(".go") and not file.endswith("_test.go"):
            check_file(os.path.join(root, file))
