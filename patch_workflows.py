import os
import glob

def patch_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    new_content = content.replace("image: iowoi/mcp-runner:latest\n", "image: iowoi/mcp-runner:latest\n      credentials:\n        username: ${{ secrets.DOCKERHUB_USERNAME }}\n        password: ${{ secrets.DOCKERHUB_TOKEN }}\n")

    with open(filepath, 'w') as f:
        f.write(new_content)

for filepath in glob.glob('.github/workflows/*.yml'):
    patch_file(filepath)
