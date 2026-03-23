#!/bin/bash
export PATH="$PATH:/usr/local/bin"
bazelisk test //ui:lint //ui:typecheck //server/... //ui:e2e_tests \
--config=remote \
--test_output=errors \
--remote_header=x-buildbuddy-api-key=vGFjlQg7X49NwoQHAfRW
