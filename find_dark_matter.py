# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0


import os

def main():
    with open("go_files.txt", "r") as f:
        files = [line.strip() for line in f.readlines()]

    source_files = {}
    test_files = set()

    for file_path in files:
        if file_path.endswith("_test.go"):
            test_files.add(file_path)
        else:
            try:
                size = os.path.getsize(file_path)
                source_files[file_path] = size
            except OSError:
                continue

    candidates = []
    for src, size in source_files.items():
        test_file = src.replace(".go", "_test.go")
        if test_file not in test_files:
            candidates.append((src, size))

    # Sort by size descending
    candidates.sort(key=lambda x: x[1], reverse=True)

    print("Top 20 potential Dark Matter candidates (large files with no tests):")
    for src, size in candidates[:20]:
        print(f"{src} ({size} bytes)")

if __name__ == "__main__":
    main()
