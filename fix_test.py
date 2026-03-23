import re

with open("server/pkg/storage/postgres/store_load_test.go", "r") as f:
    content = f.read()

content = content.replace('"github.com/stretchr/testify/assert"', '')
content = content.replace("store := NewStore(pgDB)", "store := NewStore(pgDB)\n\t\t_ = store")

with open("server/pkg/storage/postgres/store_load_test.go", "w") as f:
    f.write(content)
