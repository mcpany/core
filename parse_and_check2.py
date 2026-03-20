import os
import re

def process_file_go(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    func_pattern = re.compile(r'^func\s+(?:\([^\)]+\)\s+)?([A-Z]\w*)\s*\(', re.MULTILINE)

    missing = []

    for pattern in [func_pattern]:
        matches = pattern.finditer(content)
        for match in matches:
            name = match.group(1)
            idx = match.start()
            preceding = content[:idx].strip().split('\n')

            doc_lines = []
            for i in range(len(preceding)-1, -1, -1):
                if preceding[i].startswith('//'):
                    doc_lines.insert(0, preceding[i])
                else:
                    break

            doc_str = '\n'.join(doc_lines)

            # Wait, functions need Returns, Errors, Side Effects according to AGENTS.md right?
            if 'Summary:' not in doc_str:
                missing.append(f"{filepath} : {name}")
            else:
                if 'Returns:' not in doc_str and 'Returns' not in doc_str:
                    missing.append(f"{filepath} : {name} (Missing Returns)")
                if 'Errors:' not in doc_str and 'Errors' not in doc_str:
                    missing.append(f"{filepath} : {name} (Missing Errors)")
                if 'Side Effects:' not in doc_str and 'Side Effects' not in doc_str:
                    missing.append(f"{filepath} : {name} (Missing Side Effects)")

    return missing

all_missing = []
for root, dirs, files in os.walk('.'):
    if 'test' in root or 'vendor' in root or '.git' in root or 'ui' in root:
        continue

    for file in files:
        if file.endswith('.go') and not file.endswith('_test.go') and not 'mock' in file and not 'pb.go' in file:
            filepath = os.path.join(root, file)
            missing = process_file_go(filepath)
            all_missing.extend(missing)

print(f"Total missing: {len(all_missing)}")
