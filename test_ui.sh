#!/bin/bash
export NODE_OPTIONS="--max-old-space-size=4096"
cd ui
npm run test
