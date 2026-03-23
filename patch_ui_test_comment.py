with open("ui/tests/audit-logs.spec.ts", "r") as f:
    content = f.read()

content = content.replace("    // We mocked it so no actual file is downloaded, just checking the Toast\n", "    // We check the Toast showing successful export\n")

with open("ui/tests/audit-logs.spec.ts", "w") as f:
    f.write(content)
