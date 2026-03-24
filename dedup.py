
import sys
import re

def deduplicate_sections(filepath, header_pattern):
    with open(filepath, 'r') as f:
        content = f.read()

    # Split by the header pattern, keeping the matches
    parts = re.split(f'({header_pattern})', content)

    if len(parts) <= 1:
        return

    new_parts = [parts[0]] # Everything before the first header
    seen_headers = set()

    for i in range(1, len(parts), 2):
        header = parts[i].strip()
        body = parts[i+1]

        if header not in seen_headers:
            seen_headers.add(header)
            new_parts.append(parts[i])
            new_parts.append(body)
        else:
            # Skip this duplicate section
            pass

    with open(filepath, 'w') as f:
        f.write("".join(new_parts))

if __name__ == "__main__":
    file_to_clean = sys.argv[1]
    pattern = sys.argv[2]
    deduplicate_sections(file_to_clean, pattern)
