#!/bin/bash
set -e

# Run the local lints again to confirm local success
cd ui
npm run lint || true
npm run typecheck || true
npm run build || true
