#!/bin/bash
sed -i 's/test --local_test_jobs=1/test --local_test_jobs=2/g' .bazelrc
sed -i 's/test:remote --local_test_jobs=1/test:remote --local_test_jobs=2/g' .bazelrc
