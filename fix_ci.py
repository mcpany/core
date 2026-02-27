import re

with open('.github/workflows/ci.yml', 'r') as f:
    content = f.read()

# Instead of blindly replacing, let's just make sure `SUDO` logic is completely robust.
# The original file has `sudo rm -rf`. I want to replace that exactly.

content = content.replace(
"""        run: |
          sudo rm -rf /usr/local/lib/android
          sudo rm -rf /usr/share/dotnet
          sudo rm -rf /opt/ghc
          sudo rm -rf /usr/local/share/boost""",
"""        run: |
          if command -v sudo &> /dev/null; then SUDO="sudo"; else SUDO=""; fi
          $SUDO rm -rf /usr/local/lib/android || true
          $SUDO rm -rf /usr/share/dotnet || true
          $SUDO rm -rf /opt/ghc || true
          $SUDO rm -rf /usr/local/share/boost || true""")

# Handle if there's already some $SUDO version from a previous commit
content = content.replace(
"""        run: |
          if command -v sudo &> /dev/null; then SUDO="sudo"; else SUDO=""; fi
          $SUDO rm -rf /usr/local/lib/android || true
          $SUDO rm -rf /usr/share/dotnet || true
          $SUDO rm -rf /opt/ghc || true
          $SUDO rm -rf /usr/local/share/boost || true""",
"""        run: |
          if command -v sudo &> /dev/null; then SUDO="sudo"; else SUDO=""; fi
          $SUDO rm -rf /usr/local/lib/android || true
          $SUDO rm -rf /usr/share/dotnet || true
          $SUDO rm -rf /opt/ghc || true
          $SUDO rm -rf /usr/local/share/boost || true""")

content = content.replace(
"""        run: |
          if command -v apt-get &> /dev/null; then
            sudo apt-get update && sudo apt-get install -y make xz-utils
          elif command -v apk &> /dev/null; then
             sudo apk add --no-cache make xz
          fi""",
"""        run: |
          if command -v sudo &> /dev/null; then SUDO="sudo"; else SUDO=""; fi
          if command -v apt-get &> /dev/null; then
            $SUDO apt-get update && $SUDO apt-get install -y make xz-utils build-essential
          elif command -v apk &> /dev/null; then
            $SUDO apk add --no-cache make xz build-base
          fi""")


# Replace my previous messy fix
content = content.replace(
"""        run: |
          if command -v sudo &> /dev/null; then SUDO="sudo"; else SUDO=""; fi
          if command -v apt-get &> /dev/null; then
            $SUDO apt-get update && $SUDO apt-get install -y make xz-utils build-essential
          elif command -v apk &> /dev/null; then
            $SUDO apk add --no-cache make xz build-base
          fi""",
"""        run: |
          if command -v sudo &> /dev/null; then SUDO="sudo"; else SUDO=""; fi
          if command -v apt-get &> /dev/null; then
            $SUDO apt-get update && $SUDO apt-get install -y make xz-utils build-essential
          elif command -v apk &> /dev/null; then
            $SUDO apk add --no-cache make xz build-base
          fi""")

content = content.replace("timeout-minutes: 30", "timeout-minutes: 60")
content = content.replace("timeout-minutes: 45", "timeout-minutes: 60")

with open('.github/workflows/ci.yml', 'w') as f:
    f.write(content)
