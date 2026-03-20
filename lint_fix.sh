#!/bin/bash
set -e

# Try pre-commit on all files
pre-commit run --all-files || echo "pre-commit failed"
