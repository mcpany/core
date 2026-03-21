#!/bin/bash
cd server
for dir in pkg/* cmd/* tests/* tools/*; do
  if [ -d "$dir" ]; then
    echo "Linting $dir..."
    /app/build/env/bin/golangci-lint run "./$dir/..."
  fi
done
