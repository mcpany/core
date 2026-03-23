import re

with open("server/tests/integration/hot_reload_test.go", "r") as f:
    content = f.read()

print("File len", len(content))
