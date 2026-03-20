#!/bin/bash
cd ui
sed -i 's/ServiceTemplate.toJSON(ServiceTemplate.fromJSON(template))/template/g' tests/e2e/test-data.ts
sed -i 's/UpstreamServiceConfig.toJSON(UpstreamServiceConfig.fromJSON(service))/service/g' tests/e2e/test-data.ts
sed -i 's/User.toJSON(User.fromJSON(user))/user/g' tests/e2e/test-data.ts

sed -i '/import { ServiceTemplate }/d' tests/e2e/test-data.ts
sed -i '/import { UpstreamServiceConfig }/d' tests/e2e/test-data.ts
sed -i '/import { User }/d' tests/e2e/test-data.ts
