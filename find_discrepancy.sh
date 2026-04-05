#!/bin/bash
for f in $(find server/docs/features ui/docs/features -type f -name "*.md"); do
    echo "Processing $f"
done
