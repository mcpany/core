#!/bin/bash
export PATH=$PATH:/app/build/env/bin
cd server
bazelisk test //... > test_out.log 2>&1
