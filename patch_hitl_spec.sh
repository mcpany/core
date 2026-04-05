#!/bin/bash
sed -i "s|.locator('.bg-background\\\\/50')|.locator('[data-testid=\"hitl-card\"]')|g" ui/tests/e2e/hitl.spec.ts
