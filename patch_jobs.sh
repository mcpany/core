#!/bin/bash
echo "build:remote --local_test_jobs=1" >> .bazelrc
echo "test --local_test_jobs=1" >> .bazelrc
