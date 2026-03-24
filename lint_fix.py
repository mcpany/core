import re

with open('server/go.mod', 'r') as f:
    text = f.read()

text = text.replace("toolchain go1.26.1", "toolchain go1.24.0")

with open('server/go.mod', 'w') as f:
    f.write(text)
