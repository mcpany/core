with open("server/go.mod", "r") as f:
    content = f.read()

content = content.replace("toolchain go1.26.1", "")

with open("server/go.mod", "w") as f:
    f.write(content)
