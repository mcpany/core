import sys

def main():
    filepath = ".github/workflows/ci.yml"
    with open(filepath, "r") as f:
        lines = f.readlines()

    new_lines = []
    skip_mode = False

    # We want to remove the redundant blocks.
    # The redundancy pattern is:
    # - name: Install System Dependencies
    #   run: |
    #     sudo apt-get update || true
    #     sudo apt-get install -y xz-utils curl ca-certificates
    #
    # followed immediately by another one? Or close by.

    # Let's target the specific range or content.
    # Lines 385-394 in the previous `sed` output seem to be the ones to remove.
    # But line numbers might shift.

    # Strategy: Find the third occurrence of "Install System Dependencies" in the file.
    # The grep showed lines 210, 385, 390, 395, 479.
    # 210 is unit-test (keep).
    # 385 is e2e-parallel (duplicate 1).
    # 390 is e2e-parallel (duplicate 2).
    # 395 is e2e-parallel (keep, the detailed one).
    # 479 is build-images (keep).

    # Wait, grep line numbers are approximate if I didn't use `grep -n`. I did use `grep -n`.
    # 385 and 390 are very close.

    # Let's iterate and look for the simple block.

    simple_block = [
        "      - name: Install System Dependencies\n",
        "        run: |\n",
        "          sudo apt-get update || true\n",
        "          sudo apt-get install -y xz-utils curl ca-certificates\n",
        "\n"
    ]

    # We need to be careful not to remove the detailed block which starts similarly.
    # The detailed block continues with `echo "PATH=$PATH"` on the next line.

    i = 0
    while i < len(lines):
        # Check if we are at the start of a simple block
        if lines[i] == simple_block[0]:
            # Check if it matches the simple block pattern
            is_simple = True
            for j in range(len(simple_block)):
                if i + j >= len(lines) or lines[i+j] != simple_block[j]:
                    is_simple = False
                    break

            # If it matches, check the NEXT line to confirm it's not the detailed block
            if is_simple:
                next_line_idx = i + len(simple_block)
                if next_line_idx < len(lines):
                    next_line = lines[next_line_idx].strip()
                    # The detailed block has `echo "PATH=$PATH"` as the next instruction in the run block?
                    # No, the run block continues.
                    # Wait, the simple block in my list ends with "\n".
                    # In YAML, `run: |` continues until indentation changes.

                    # Let's look at the file content again.
                    # Block 1:
                    # run: |
                    #   sudo apt-get update || true
                    #   sudo apt-get install -y xz-utils curl ca-certificates
                    # <empty line or next step>

                    # Block 3:
                    # run: |
                    #   sudo apt-get update || true
                    #   sudo apt-get install -y xz-utils curl ca-certificates
                    #   echo "PATH=$PATH"

                    # So if line i+4 starts with indentation and content, it's the detailed block.
                    # simple_block definition above assumes 4 lines + empty line.

                    # Let's refine the matching logic.
                    pass

    # New strategy:
    # Identify the lines 385-394 roughly.
    # Read the file into a list.
    # Find the sequence of duplicates in `e2e-parallel`.

    # Finding `e2e-parallel` job start
    try:
        e2e_idx = -1
        for idx, line in enumerate(lines):
            if "e2e-parallel:" in line:
                e2e_idx = idx
                break

        if e2e_idx == -1:
            print("Could not find e2e-parallel job")
            return

        # Now search for duplicates within e2e-parallel steps
        # We look for adjacent blocks of "Install System Dependencies"

        # We will iterate through lines after e2e_idx
        i = e2e_idx
        while i < len(lines):
            if "name: Install System Dependencies" in lines[i]:
                # Found a block. Check its content.
                # If it's the short version, and followed immediately by another "Install System Dependencies", remove it.

                # Check checking content length
                # Short version has `run: |` and 2 command lines.
                # Long version has `run: |` and many command lines.

                # Let's count indentation lines.
                j = i + 2 # skip name and run
                cmd_count = 0
                while j < len(lines) and (lines[j].strip() == "" or lines[j].startswith("          ")): # 10 spaces
                    if lines[j].strip() != "":
                        cmd_count += 1
                    j += 1

                # Short block has 2 commands.
                # Long block has ~15 commands.

                if cmd_count == 2:
                    # This is a short block.
                    # Check if the next non-empty line starts a new "Install System Dependencies" step.
                    k = j
                    while k < len(lines) and lines[k].strip() == "":
                        k += 1

                    if k < len(lines) and "name: Install System Dependencies" in lines[k]:
                        # It is a duplicate! Remove lines i to k-1 (or just i to j, handling empty lines)
                        # Actually we can just skip adding these lines to new_lines.
                        print(f"Removing duplicate block at line {i+1}")
                        i = k # Skip to the next block
                        continue

            new_lines.append(lines[i])
            i += 1

    except Exception as e:
        print(f"Error: {e}")
        return

    with open(filepath, "w") as f:
        f.writelines(new_lines)

if __name__ == "__main__":
    main()
