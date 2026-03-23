import re

with open("server/pkg/storage/postgres/store_load_test.go", "r") as f:
    content = f.read()

content = content.replace('\n\t"github.com/stretchr/testify/require"\n', '"github.com/stretchr/testify/require"\n')
content = content.replace('\t\t\tdb, mock, err := sqlmock.New()', '\t\tdb, mock, err := sqlmock.New()')

with open("server/pkg/storage/postgres/store_load_test.go", "w") as f:
    f.write(content)
