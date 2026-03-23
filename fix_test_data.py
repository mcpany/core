# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import os

filepath = "ui/tests/e2e/test-data.ts"
with open(filepath, "r") as f:
    content = f.read()

content = content.replace("import { ServiceTemplate } from '../../../proto/config/v1/service_template';", "")
content = content.replace("import { UpstreamServiceConfig } from '../../../proto/config/v1/upstream_service';", "")
content = content.replace("import { User } from '../../../proto/config/v1/user';", "")

content = content.replace("UpstreamServiceConfig.toJSON(UpstreamServiceConfig.fromJSON(service))", "service")
content = content.replace("ServiceTemplate.toJSON(ServiceTemplate.fromJSON(template))", "template")
content = content.replace("User.toJSON(User.fromJSON(user))", "user")

with open(filepath, "w") as f:
    f.write(content)
