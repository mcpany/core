import re

filename = 'server/pkg/upstream/sql/tool.go'

with open(filename, 'r') as f:
    content = f.read()

content = content.replace('req *ExecutionRequest', 'req *tool.ExecutionRequest')

with open(filename, 'w') as f:
    f.write(content)
