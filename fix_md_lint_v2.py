import sys
import re

def fix_content(content):
    lines = content.splitlines()
    new_lines = []

    for i, line in enumerate(lines):
        # Rule MD022: Headers should be surrounded by blank lines
        if line.startswith('#'):
            # Blank line before
            if i > 0 and new_lines and new_lines[-1].strip() != '':
                new_lines.append('')

            new_lines.append(line)

            # Blank line after
            if i < len(lines) - 1:
                next_line = lines[i+1]
                if next_line.strip() != '' and not next_line.startswith('#'):
                    new_lines.append('')
            continue

        # Rule MD032: Lists should be surrounded by blank lines
        # This handles blank line BEFORE a list
        is_list_item = re.match(r'^\s*([-*+]|\d+\.)\s', line)
        if is_list_item:
            if i > 0 and new_lines and new_lines[-1].strip() != '' and not re.match(r'^\s*([-*+]|\d+\.)\s', new_lines[-1]) and not new_lines[-1].startswith('#'):
                new_lines.append('')

        new_lines.append(line)

        # This handles blank line AFTER a list
        # If current line is a list item and next line is not a list item, not a header, and not empty
        if is_list_item and i < len(lines) - 1:
            next_line = lines[i+1]
            if next_line.strip() != '' and not next_line.startswith('#') and not re.match(r'^\s*([-*+]|\d+\.)\s', next_line):
                 new_lines.append('')

    return '\n'.join(new_lines) + ('\n' if content.endswith('\n') else '')

if __name__ == "__main__":
    for filepath in sys.argv[1:]:
        with open(filepath, 'r') as f:
            content = f.read()
        fixed = fix_content(content)
        # Remove multiple blank lines
        fixed = re.sub(r'\n{3,}', '\n\n', fixed)

        with open(filepath, 'w') as f:
            f.write(fixed)
