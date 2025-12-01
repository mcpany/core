#!/bin/bash

set -e

echo "Verifying 'Makefile' for 'test-fast' configuration..."

if grep -q "test-fast:" "Makefile" && grep -q -- "-tags=!e2e" "Makefile"; then
  echo "Success: 'Makefile' is correctly configured to exclude e2e tests from 'test-fast'."
  exit 0
else
  echo "Error: 'Makefile' is not correctly configured. 'test-fast' should not run e2e tests."
  exit 1
fi
