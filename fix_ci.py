import re

with open(".github/workflows/ci.yml", "r") as f:
    content = f.read()

# 1. Remove Generate Protobufs step from test jobs, BUT NOT FROM build-images
# build-images step has "make gen" inside a run: | block.
# Test jobs have:
#       - name: Generate Protobufs
#         run: make gen

content = re.sub(r'      - name: Generate Protobufs\n        run: make gen\n\n', '', content)

# 2. Remove apt-get install dependencies blocks
apt_block = r"""      - name: Install dependencies
        run: \|
          if command -v apt-get &> /dev/null; then
            sudo apt-get update && sudo apt-get install -y make xz-utils.*?
          elif command -v apk &> /dev/null; then
             sudo apk add --no-cache make xz.*?
          fi\n\n"""

content = re.sub(apt_block, '', content, flags=re.DOTALL)

with open(".github/workflows/ci.yml", "w") as f:
    f.write(content)
