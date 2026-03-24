#!/bin/bash
set -e

echo "Simulating CI lint environment..."

# Pretend we are in venv
export PATH=$PWD/venv/bin:$PWD/build/env/bin:$PATH

echo "Running make lint..."
make lint
echo "exit code: $?"
