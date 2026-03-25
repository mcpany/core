import re

with open("pkg/tool/callable_test.go", "r") as f:
    content = f.read()

content = content.replace('toolDef := &configv1.ToolDefinition{\n\t\tName: "test_tool",\n\t}', 'toolDef := configv1.ToolDefinition_builder{Name: proto.String("test_tool")}.Build()')
content = content.replace('toolDef := &configv1.ToolDefinition{Name: "test_tool"}', 'toolDef := configv1.ToolDefinition_builder{Name: proto.String("test_tool")}.Build()')

if '"google.golang.org/protobuf/proto"' not in content:
    content = content.replace('import (', 'import (\n\t"google.golang.org/protobuf/proto"\n')

with open("pkg/tool/callable_test.go", "w") as f:
    f.write(content)
