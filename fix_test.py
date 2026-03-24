import re

with open("server/pkg/storage/postgres/store_load_test.go", "r") as f:
    content = f.read()

content = content.replace('\tconfigv1 "github.com/mcpany/core/proto/config/v1""github.com/stretchr/testify/require"', '\tconfigv1 "github.com/mcpany/core/proto/config/v1"\n\t"github.com/stretchr/testify/require"')
with open("server/pkg/storage/postgres/store_load_test.go", "w") as f:
    f.write(content)
