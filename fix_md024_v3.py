import sys
import re

def fix_file(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()

    seen = {}
    new_lines = []

    for line in lines:
        match_header = re.match(r'^(#+)\s+(.+)$', line)
        if match_header:
            level = match_header.group(1)
            title = match_header.group(2).strip()
            # Track duplicates across all levels to satisfy strict MD024
            key = title
            if key in seen:
                seen[key] += 1
                # Contextualize
                date_match = re.search(r'\d{4}-\d{2}-\d{2}', title)
                if date_match:
                    date = date_match.group(0)
                    new_title = title.replace(date, f"{date} Phase {seen[key]}")
                else:
                    new_title = f"{title} (Part {seen[key]})"
                new_lines.append(f"{level} {new_title}\n")
            else:
                seen[key] = 1
                new_lines.append(line)
        else:
            new_lines.append(line)

    with open(filepath, 'w') as f:
        f.writelines(new_lines)

if __name__ == "__main__":
    for arg in sys.argv[1:]:
        fix_file(arg)
