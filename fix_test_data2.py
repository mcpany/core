import re

with open("ui/tests/e2e/test-data.ts", "r") as f:
    content = f.read()

# Replace map usage with just object to avoid protoc dependencies in playwright since they don't compile locally outside bazel properly without full setup.
content = content.replace("import { ServiceTemplate } from '../../../proto/config/v1/service_template';", "// import { ServiceTemplate } from '../../../proto/config/v1/service_template';")
content = content.replace("import { UpstreamServiceConfig } from '../../../proto/config/v1/upstream_service';", "// import { UpstreamServiceConfig } from '../../../proto/config/v1/upstream_service';")
content = content.replace("import { User } from '../../../proto/config/v1/user';", "// import { User } from '../../../proto/config/v1/user';")

content = content.replace("].map((service) => UpstreamServiceConfig.toJSON(UpstreamServiceConfig.fromJSON(service)));", "] as any;")
content = content.replace("].map((template) => ServiceTemplate.toJSON(ServiceTemplate.fromJSON(template)));", "] as any;")
content = content.replace("].map((user) => User.toJSON(User.fromJSON(user)));", "] as any;")

with open("ui/tests/e2e/test-data.ts", "w") as f:
    f.write(content)
