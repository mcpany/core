import re

with open(".github/workflows/ci.yml", "r") as f:
    content = f.read()

# 1. Switch to ubuntu:22.04 in build-images
content = content.replace("public.ecr.aws/docker/library/ubuntu:24.04", "public.ecr.aws/docker/library/ubuntu:22.04")

# 2. Update the apt-get script block in all jobs
# We look for the block we just wrote:
# export DEBIAN_FRONTEND=noninteractive
#  rm -rf /var/lib/apt/lists/*
# for i in 1 2 3 4 5; do
#   ( apt-get update) && break || {
#     echo "apt-get update failed, retrying (/5)..."
#     sleep 5
#   }
# done
#  apt-get install -y --no-install-recommends ...

# We want to replace it with:
# export DEBIAN_FRONTEND=noninteractive
# # Nuke everything
#  rm -rf /var/lib/apt/lists/*
#  rm -f /var/lib/dpkg/lock-frontend /var/lib/dpkg/lock
# # Force IPv4 and retry
# for i in 1 2 3 4 5; do
#   ( apt-get -o Acquire::ForceIPv4=true update) && break || {
#     echo "apt-get update failed, retrying (/5)..."
#     sleep 5
#   }
# done
# # Configure if interrupted
#  dpkg --configure -a || true
#  apt-get -o Acquire::ForceIPv4=true install -y --no-install-recommends ...

# The install command varies (git/make/curl/unzip vs make/xz-utils...).
# So we need to match the block up to the install command.

# Regex to match the block:
block_pattern = r"(export DEBIAN_FRONTEND=noninteractive\n\s+$SUDO rm -rf /var/lib/apt/lists/\*\n\s+for i in 1 2 3 4 5; do\n\s+\($SUDO apt-get update\) && break \|\| \{\n\s+echo \"apt-get update failed, retrying \($i/5\)\.\.\.\"\n\s+sleep 5\n\s+\}\n\s+done\n\s+$SUDO apt-get install -y --no-install-recommends)"

replacement = r"""export DEBIAN_FRONTEND=noninteractive
             rm -rf /var/lib/apt/lists/*
             rm -f /var/lib/dpkg/lock-frontend /var/lib/dpkg/lock
            for i in 1 2 3 4 5; do
              ( apt-get -o Acquire::ForceIPv4=true update) && break || {
                echo "apt-get update failed, retrying (/5)..."
                sleep 5
              }
            done
             dpkg --configure -a || true
             apt-get -o Acquire::ForceIPv4=true install -y --no-install-recommends"""

# We need to handle indentation.
# The  doesn't automatically handle indentation if we provide a multiline string.
# But  accepts a function.

def replace_block(match):
    # We don't capture indentation in the group, so we assume the indentation of the matched string is correct relative to itself?
    # No, the regex matches explicitly indented strings?
    # No, in the previous file write I used spaces.
    # Let's just use string replace for the whole block if possible.
    # But indentation varies? No, I copy pasted it.
    # Indentation is 12 spaces usually.
    return match.group(0).replace("apt-get update", "apt-get -o Acquire::ForceIPv4=true update").replace(
        "rm -rf /var/lib/apt/lists/*",
        "rm -rf /var/lib/apt/lists/*\n             rm -f /var/lib/dpkg/lock-frontend /var/lib/dpkg/lock"
    ).replace(
        "done\n             apt-get install",
        "done\n             dpkg --configure -a || true\n             apt-get -o Acquire::ForceIPv4=true install"
    )

# Actually, I can just use a regex that matches the logic and replaces it.
# The logic is:
# 1. rm -rf lists
# 2. for i in 1..5 update
# 3. install

# I'll just write the file again using the known structure, to be safe.
# It's tedious but safe.
pass

with open(".github/workflows/ci.yml", "w") as f:
    f.write(content)
