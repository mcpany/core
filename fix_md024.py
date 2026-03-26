import sys
import re

def fix_file(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    seen = {}
    new_lines = []

    for line in lines:
        if line.startswith('## '):
            header = line.strip()
            if header in seen:
                seen[header] += 1
                # Contextualize based on date in file if possible or just append index
                # Looking for YYYY-MM-DD
                match = re.search(r'\d{4}-\d{2}-\d{2}', header)
                if match:
                    date = match.group(0)
                    new_header = header.replace(date, f"{date} ({seen[header]})")
                else:
                    new_header = f"{header} ({seen[header]})"
                new_lines.append(new_header + '\n')
            else:
                seen[header] = 1
                new_lines.append(line)
        else:
            new_lines.append(line)

    with open(filepath, 'w') as f:
        f.writelines(new_lines)

if __name__ == "__main__":
    for arg in sys.argv[1:]:
        fix_file(arg)
