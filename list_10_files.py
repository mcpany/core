# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

import random
import os

docs = []
for root, _, files in os.walk('ui/docs'):
    for file in files:
        if file.endswith('.md'):
            docs.append(os.path.join(root, file))

for root, _, files in os.walk('server/docs'):
    for file in files:
        if file.endswith('.md'):
            docs.append(os.path.join(root, file))

random.seed(42)
selected = random.sample(docs, 10)
print("\n".join(selected))
