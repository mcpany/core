import sys

def main():
    filepath = ".github/workflows/ci.yml"
    with open(filepath, "r") as f:
        content = f.read()

    # 1. Update Install dependencies to include curl and unzip
    # This block appears in multiple jobs running on oci-runner
    old_dep = """      - name: Install dependencies
        run: |
          if command -v apt-get &> /dev/null; then
            sudo apt-get update && sudo apt-get install -y make xz-utils
          elif command -v apk &> /dev/null; then
             sudo apk add --no-cache make xz
          fi"""

    new_dep = """      - name: Install dependencies
        run: |
          SUDO=""
          if command -v sudo >/dev/null; then
            SUDO="sudo"
          fi
          if command -v apt-get &> /dev/null; then
            $SUDO apt-get update && $SUDO apt-get install -y make xz-utils build-essential curl unzip
          elif command -v apk &> /dev/null; then
             $SUDO apk add --no-cache make xz build-base curl unzip
          fi"""

    # Need to handle the previous fix attempt if it's there
    prev_fix_dep = """      - name: Install dependencies
        run: |
          SUDO=""
          if command -v sudo >/dev/null; then
            SUDO="sudo"
          fi
          if command -v apt-get &> /dev/null; then
            $SUDO apt-get update && $SUDO apt-get install -y make xz-utils build-essential
          elif command -v apk &> /dev/null; then
             $SUDO apk add --no-cache make xz build-base
          fi"""

    if prev_fix_dep in content:
        content = content.replace(prev_fix_dep, new_dep)
    elif old_dep in content:
        content = content.replace(old_dep, new_dep)
    else:
        print("Warning: Could not find dependency install block to replace.")

    # 2. Fix build-images job - it uses a different install dependencies block or none?
    # build-images uses free disk space then checkout. It does NOT have an "Install dependencies" step.
    # It relies on 'make prepare' inside the runner.
    # However, 'make prepare' needs curl/unzip.
    # We should ADD the install dependencies step to build-images before "Set up Go" or "Generate Protobufs"

    build_images_marker = """      - uses: actions/checkout@v6

      - name: Set up Go"""

    build_images_install = """      - uses: actions/checkout@v6

      - name: Install Dependencies
        run: |
          SUDO=""
          if command -v sudo >/dev/null; then
            SUDO="sudo"
          fi
          if command -v apt-get &> /dev/null; then
            $SUDO apt-get update && $SUDO apt-get install -y make xz-utils build-essential curl unzip
          elif command -v apk &> /dev/null; then
             $SUDO apk add --no-cache make xz build-base curl unzip
          fi

      - name: Set up Go"""

    if build_images_marker in content:
        content = content.replace(build_images_marker, build_images_install)
    else:
        print("Warning: Could not find build-images marker.")

    # 3. Increase ui-test timeout (already done, but check)
    # 4. Increase k8s-e2e timeout (already done, but check)

    with open(filepath, "w") as f:
        f.write(content)

if __name__ == "__main__":
    main()
