import os

def check_file(filepath):
    print(f"Checking {filepath}...")
    try:
        with open(filepath, 'rb') as f:
            content = f.read()
    except Exception as e:
        print(f"  Error reading: {e}")
        return

    # Remove common non-ASCII
    final_content = content.replace(b'\xe2\x80\x93', b'-').replace(b'\xe2\x80\x94', b'-')
    final_content = final_content.replace(b'\xe2\x80\x9c', b'"').replace(b'\xe2\x80\x9d', b'"')
    final_content = final_content.replace(b'\xe2\x80\x98', b"'").replace(b'\xe2\x80\x99', b"'")

    # Remove remaining non-ASCII
    final_content = bytearray(b for b in final_content if b <= 127)

    # Ensure single newline at end and no trailing whitespace
    try:
        content_str = final_content.decode('ascii')
        lines = content_str.splitlines()
        new_lines = [line.rstrip() for line in lines]
        final_str = "\n".join(new_lines).rstrip() + "\n"

        with open(filepath, 'w', encoding='ascii') as f:
            f.write(final_str)
        print(f"  Processed {filepath}")
    except Exception as e:
        print(f"  Error processing {filepath}: {e}")

files = [
    "docs/02_strategic_vision.md",
    "docs/03_feature_inventory.md",
    "docs/features/design-ari-hub.md",
    "docs/research/market-sync-2026-06-14.md",
    "server/roadmap.md",
    "ui/roadmap.md",
    ".circleci/config.yml",
    "MODULE.bazel"
]

for f in files:
    if os.path.exists(f):
        check_file(f)
