#!/bin/bash
# Re-run build to capture the exact failure from e2e tests.
bazelisk test //k8s/operator/tests:e2e_test
