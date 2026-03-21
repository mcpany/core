import os
import re

file_path = "server/pkg/app/api_users_extra_test.go"
with open(file_path, "r") as f:
    content = f.read()

content = content.replace('configv1.User_builder{Id: proto.String("existing-user")}.Build()', '&configv1.User{Id: proto.String("existing-user")}')
content = content.replace('configv1.User_builder{Id: proto.String("user-test-detail")}.Build()', '&configv1.User{Id: proto.String("user-test-detail")}')

with open(file_path, "w") as f:
    f.write(content)
