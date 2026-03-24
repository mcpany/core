import os

def purge_file(filepath):
    print(f"Purging {filepath}...")
    try:
        with open(filepath, 'rb') as f:
            content = f.read()
    except Exception as e:
        print(f"  Error: {e}")
        return

    # Replace common non-ASCII
    content = content.replace(b'\xe2\x80\x93', b'-')
    content = content.replace(b'\xe2\x80\x94', b'-')
    content = content.replace(b'\xe2\x80\x9c', b'"')
    content = content.replace(b'\xe2\x80\x9d', b'"')
    content = content.replace(b'\xe2\x80\x98', b"'")
    content = content.replace(b'\xe2\x80\x99', b"'")
    content = content.replace(b'\xe2\x80\xa2', b'*')

    # Strip any remaining non-ASCII
    final_content = bytearray(b for b in content if 32 <= b <= 126 or b in (9, 10, 13))

    # Standardize newlines (single newline at EOF, no trailing whitespace)
    try:
        text = final_content.decode('ascii')
        lines = [line.rstrip() for line in text.splitlines()]
        final_text = "\n".join(lines).rstrip() + "\n"
        with open(filepath, 'w', encoding='ascii') as f:
            f.write(final_text)
        print(f"  Success.")
    except Exception as e:
        print(f"  Final write error: {e}")

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
        purge_file(f)
