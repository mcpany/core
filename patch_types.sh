#!/bin/bash
cd ui
find tests/e2e -type f -name "test-data.ts" -exec sed -i 's/ServiceTemplate/any/g' {} +
find tests/e2e -type f -name "test-data.ts" -exec sed -i 's/UpstreamServiceConfig/any/g' {} +
find tests/e2e -type f -name "test-data.ts" -exec sed -i 's/User/any/g' {} +
