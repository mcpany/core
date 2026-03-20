import os
import re

def clean_content(content):
    # Standardize date format to [YYYY-MM-DD]
    content = re.sub(r'(?<!\[)(\d{4}-\d{2}-\d{2})(?!\])', r'[\1]', content)
    # Strip non-ASCII
    content = "".join(c for c in content if ord(c) < 128)
    return content

targets = [
    'docs/02_strategic_vision.md',
    'docs/03_feature_inventory.md',
    'server/roadmap.md',
    'ui/roadmap.md',
    'docs/features/design-dag-middleware.md',
    'docs/features/design-rgi-provider.md',
    'docs/research/market-sync-2026-06-18.md'
]

for filepath in targets:
    if os.path.exists(filepath):
        with open(filepath, 'r') as f:
            c = f.read()
        cleaned = clean_content(c)
        if cleaned != c:
            with open(filepath, 'w') as f:
                f.write(cleaned)
            print(f"Sanitized: {filepath}")
