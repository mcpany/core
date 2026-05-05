#!/bin/bash
# Check all docs
for f in $(find ui/docs server/docs -type f -name "*.md"); do
    echo "Processing $f"
done
