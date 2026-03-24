import re

with open("server/pkg/storage/postgres/store_load_test.go", "r") as f:
    content = f.read()

content = content.replace('\n\t"github.com/stretchr/testify/assert"\n\t"github.com/stretchr/testify/require"\n', '\n\t"github.com/stretchr/testify/require"\n')

with open("server/pkg/storage/postgres/store_load_test.go", "w") as f:
    f.write(content)
