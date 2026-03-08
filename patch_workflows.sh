#!/bin/bash
for file in .github/workflows/*.yml; do
  sed -i 's|image: iowoi/mcp-runner:latest|image: iowoi/mcp-runner:latest\n      credentials:\n        username: ${{ secrets.DOCKERHUB_USERNAME }}\n        password: ${{ secrets.DOCKERHUB_TOKEN }}|g' "$file"
done
