import os
import re

def process_go_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()

    out_lines = []
    i = 0
    while i < len(lines):
        line = lines[i]

        # Match only func declarations
        match_func = re.match(r'^func\s+(?:(?:\([^)]+\)\s+)|)([A-Z]\w*)', line)
        if match_func:
            name = match_func.group(1)

            # Find the preceding comment block
            j = len(out_lines) - 1
            has_comment = False
            comment_start = -1
            while j >= 0 and out_lines[j].strip().startswith('//'):
                has_comment = True
                comment_start = j
                j -= 1

            if has_comment:
                comment_block = out_lines[comment_start:]
                full_comment = ''.join(comment_block)

                # Check structure elements
                new_additions = []

                sig = line.split('{', 1)[0].strip()
                # A very basic check for has args and has returns
                has_args = not sig.split('(', 1)[1].startswith(')')
                has_returns = ')' in sig and sig.rsplit(')', 1)[1].strip() != ''

                if has_args and 'Parameters:' not in full_comment:
                    new_additions.append('//\n// Parameters:\n//   - None described.\n')

                if has_returns and 'Returns:' not in full_comment:
                    new_additions.append('//\n// Returns:\n//   - None described.\n')

                if 'Errors:' not in full_comment:
                    new_additions.append('//\n// Errors:\n//   - None specified.\n')

                if 'Side Effects:' not in full_comment:
                    new_additions.append('//\n// Side Effects:\n//   - None.\n')

                if new_additions:
                    # Insert right before the current line
                    out_lines.extend(new_additions)

        out_lines.append(line)
        i += 1

    with open(filepath, 'w', encoding='utf-8') as f:
        f.writelines(out_lines)

for root, dirs, files in os.walk('server/pkg'):
    for file in files:
        if file.endswith('.go') and not file.endswith('_test.go'):
            process_go_file(os.path.join(root, file))

for root, dirs, files in os.walk('server/cmd'):
    for file in files:
        if file.endswith('.go') and not file.endswith('_test.go'):
            process_go_file(os.path.join(root, file))
