import os

def check_file(filepath):
    print(f"Checking {filepath}...")
    try:
        with open(filepath, 'rb') as f:
            content = f.read()
    except Exception as e:
        print(f"  Error reading: {e}")
        return

    # Check for non-ASCII
    non_ascii = [ (i, b) for i, b in enumerate(content) if b > 127 ]
    if non_ascii:
        print(f"  Found {len(non_ascii)} non-ASCII bytes")
        new_content = bytearray()
        for b in content:
            if b <= 127:
                new_content.append(b)
            else:
                # Replace common non-ASCII with ASCII
                if b == 0xE2: # Part of smart quotes/dashes usually
                    pass # We'll just skip or handle sequences if we were fancy
                else:
                    new_content.append(ord('?'))

        # Simple strip of common UTF-8 sequences for em-dash (0xE2 0x80 0x93/0x94)
        final_content = content.replace(b'\xe2\x80\x93', b'-').replace(b'\xe2\x80\x94', b'-')
        final_content = final_content.replace(b'\xe2\x80\x9c', b'"').replace(b'\xe2\x80\x9d', b'"')
        final_content = final_content.replace(b'\xe2\x80\x98', b"'").replace(b'\xe2\x80\x99', b"'")

        # Remove remaining non-ASCII
        final_content = bytearray(b for b in final_content if b <= 127)

        with open(filepath, 'wb') as f:
            f.write(final_content)
        print(f"  Cleaned {filepath}")

    # Check for trailing whitespace
    with open(filepath, 'r', encoding='ascii', errors='ignore') as f:
        lines = f.readlines()

    new_lines = [line.rstrip() + '\n' for line in lines]

    # Ensure single newline at end
    content_str = "".join(new_lines).rstrip() + "\n"

    with open(filepath, 'w', encoding='ascii') as f:
        f.write(content_str)

files = [
    "docs/02_strategic_vision.md",
    "docs/03_feature_inventory.md",
    "docs/features/design-ari-hub.md",
    "docs/research/market-sync-2026-06-14.md",
    "server/roadmap.md",
    "ui/roadmap.md",
    "MODULE.bazel"
]

for f in files:
    if os.path.exists(f):
        check_file(f)
