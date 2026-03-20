import sys

def dedupe_section(filename, marker):
    with open(filename, 'r') as f:
        content = f.read()

    parts = content.split(marker)
    if len(parts) <= 2:
        return # No duplication or only one occurrence

    # We want to keep the first occurrence of the marker and its associated content,
    # but only if the content is different. If the content is identical, we remove the duplicate.
    # Actually, the user says I appended it twice.
    # So parts[0] is prefix, parts[1] is first duplicate, parts[2] is second duplicate, etc.

    # Simple fix: Keep parts[0] + marker + parts[1] + "".join(parts[2:] if they are not just duplicates of parts[1])
    # But wait, parts[2:] might contain the rest of the file.

    # Let's assume the duplication happened at the top or bottom.
    # If I did: content = new_stuff + content, and new_stuff was marker + info
    # Then parts[0] is empty. parts[1] is info. parts[2] is info + rest of file? No.

    # Let's just remove the first occurrence if it is identical to the second.
    if parts[1].strip() == parts[2].split('\n\n')[0].strip(): # very specific check
         # This is getting complicated. Let's just use the fact that I know I appended twice.
         pass

# Better approach: find the exact duplicated string and replace it with a single instance.
def fix_file(filename):
    with open(filename, 'r') as f:
        content = f.read()

    # Check for large duplicated blocks
    lines = content.splitlines()
    n = len(lines)
    for length in range(n // 2, 5, -1):
        for i in range(n - 2 * length + 1):
            if lines[i:i+length] == lines[i+length:i+2*length]:
                print(f"Found duplicate in {filename} of length {length} at line {i}")
                new_lines = lines[:i+length] + lines[i+2*length:]
                with open(filename, 'w') as f2:
                    f2.write('\n'.join(new_lines) + '\n')
                return True
    return False

fix_file("docs/02_strategic_vision.md")
fix_file("docs/03_feature_inventory.md")
fix_file("server/roadmap.md")
fix_file("ui/roadmap.md")
