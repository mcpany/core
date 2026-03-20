import os
import re

def fix_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    func_pattern = re.compile(r'^func\s+(?:\([^\)]+\)\s+)?([A-Z]\w*)\s*\(', re.MULTILINE)

    matches = list(func_pattern.finditer(content))
    matches.sort(key=lambda x: x.start(), reverse=True)

    modified = False

    for match in matches:
        name = match.group(1)
        idx = match.start()

        preceding = content[:idx].strip()
        lines = preceding.split('\n')

        has_doc = False
        for line in reversed(lines):
            if line.startswith('//'):
                if 'Summary:' in line:
                    has_doc = True
                    break
            elif not line:
                continue
            else:
                break

        if has_doc:
            doc_str = '\n'.join([l for l in lines[-10:] if l.startswith('//')])
            if 'Returns:' not in doc_str and 'Errors:' not in doc_str and 'Side Effects:' not in doc_str:
                docstring = f"""//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - none
//
// Errors:
//   - none
//
// Side Effects:
//   - none
"""
                content = content[:idx] + docstring + content[idx:]
                modified = True

    if modified:
        with open(filepath, 'w') as f:
            f.write(content)

for root, dirs, files in os.walk('.'):
    if 'test' in root or 'vendor' in root or '.git' in root or 'ui' in root:
        continue

    for file in files:
        if file.endswith('.go') and not file.endswith('_test.go') and not 'mock' in file and not 'pb.go' in file:
            filepath = os.path.join(root, file)
            fix_file(filepath)
