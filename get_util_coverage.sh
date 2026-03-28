#!/bin/bash
./bazelisk coverage //server/pkg/util:util_test
ls -l bazel-out/_coverage/_coverage_report.dat
