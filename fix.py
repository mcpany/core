import sys

for filepath in sys.argv[1:]:
    with open(filepath, "r") as f:
        content = f.read()

    # 1-indent
    content = content.replace(
        "\tvar buf []byte\n\tbuf = append(buf, '[')",
        "\t// ⚡ BOLT: Pre-allocate buffer to prevent O(N) reallocations during JSON array construction.\n\t// Randomized Selection from Top 5 High-Impact Targets.\n\tbuf := make([]byte, 0, 1024)\n\tbuf = append(buf, '[')"
    )

    # 3-indent
    content = content.replace(
        "\t\t\tvar buf []byte\n\t\t\tbuf = append(buf, '[')",
        "\t\t\t// ⚡ BOLT: Pre-allocate buffer to prevent O(N) reallocations during JSON array construction.\n\t\t\t// Randomized Selection from Top 5 High-Impact Targets.\n\t\t\tbuf := make([]byte, 0, 1024)\n\t\t\tbuf = append(buf, '[')"
    )

    with open(filepath, "w") as f:
        f.write(content)
