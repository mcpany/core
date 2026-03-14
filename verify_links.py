import os
import re

def check_links():
    docs_dir = 'docs'
    all_files = []
    for root, dirs, files in os.walk(docs_dir):
        for file in files:
            if file.endswith('.md'):
                all_files.append(os.path.join(root, file))

    # Also check roadmaps
    all_files.append('server/roadmap.md')
    all_files.append('ui/roadmap.md')

    for filepath in all_files:
        with open(filepath, 'r') as f:
            content = f.read()
            # Find markdown links [text](path)
            links = re.findall(r'\[.*?\]\((.*?)\)', content)
            for link in links:
                if link.startswith('http') or link.startswith('#'):
                    continue

                # Normalize path
                link_path = os.path.join(os.path.dirname(filepath), link)
                link_path = link_path.split('#')[0] # remove anchors

                if not os.path.exists(link_path):
                    print(f"Broken link in {filepath}: {link} (resolved to {link_path})")

if __name__ == "__main__":
    check_links()
