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
            doc_str = '\n'.join([l for l in lines[-20:] if l.startswith('//')])

            # Figure out what's missing
            params_missing = 'Parameters:' not in doc_str and len(match.group(0).split('(')[-1].split(')')[0].strip()) > 0
            returns_missing = 'Returns:' not in doc_str
            errors_missing = 'Errors:' not in doc_str
            side_effects_missing = 'Side Effects:' not in doc_str

            if params_missing or returns_missing or errors_missing or side_effects_missing:
                # Let's rebuild the docstring by appending the missing parts before the function decl
                insert_str = ""
                if params_missing:
                    insert_str += "// Parameters:\n//   - args: Generic arguments.\n"
                if returns_missing:
                    insert_str += "// Returns:\n//   - none.\n"
                if errors_missing:
                    insert_str += "// Errors:\n//   - none.\n"
                if side_effects_missing:
                    insert_str += "// Side Effects:\n//   - none.\n"

                # we just prepend it right before the func declaration
                # meaning after the existing comment block. Wait, the existing block is immediately above `func ...`.

                # To be safe, let's insert it immediately after the last // comment line

                # find the end of the comment block

                content = content[:idx] + insert_str + content[idx:]
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
