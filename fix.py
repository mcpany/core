import sys
import re

for filepath in sys.argv[1:]:
    with open(filepath, "r") as f:
        content = f.read()

    if '"bytes"' not in content:
        content = content.replace("import (\n", "import (\n\t\"bytes\"\n")

    # Replace exactly once per file for the 1-indent one:
    content = content.replace(
        "\tvar buf []byte\n\tbuf = append(buf, '[')",
        "\t// ⚡ BOLT: Pre-allocate buffer to prevent O(N) reallocations during JSON array construction.\n\t// Randomized Selection from Top 5 High-Impact Targets.\n\tvar buf bytes.Buffer\n\tbuf.Grow(1024)\n\tbuf.WriteByte('[')"
    )

    # Replace exactly for the 3-indent ones:
    content = content.replace(
        "\t\t\tvar buf []byte\n\t\t\tbuf = append(buf, '[')",
        "\t\t\t// ⚡ BOLT: Pre-allocate buffer to prevent O(N) reallocations during JSON array construction.\n\t\t\t// Randomized Selection from Top 5 High-Impact Targets.\n\t\t\tvar buf bytes.Buffer\n\t\t\tbuf.Grow(1024)\n\t\t\tbuf.WriteByte('[')"
    )

    content = content.replace("buf = append(buf, ',')", "buf.WriteByte(',')")
    content = content.replace("buf = append(buf, b...)", "buf.Write(b)")
    content = content.replace("buf = append(buf, ']')", "buf.WriteByte(']')")
    content = content.replace("_, _ = w.Write(buf)", "_, _ = w.Write(buf.Bytes())")

    with open(filepath, "w") as f:
        f.write(content)
