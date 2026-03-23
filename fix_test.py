import re

with open("server/pkg/storage/postgres/store_load_test.go", "r") as f:
    content = f.read()

content = content.replace("github.com/stretchr/testify/require\n\t\"github.com/stretchr/testify/assert\"", "github.com/stretchr/testify/require")

with open("server/pkg/storage/postgres/store_load_test.go", "w") as f:
    f.write(content)
