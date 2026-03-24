#!/bin/bash
export PATH="$PATH:/usr/local/bin"
bazelisk test //server/tests/integration/... //ui/... \
--config=remote \
--test_output=errors \
--remote_header=x-buildbuddy-api-key=vGFjlQg7X49NwoQHAfRW \
--jobs=2
