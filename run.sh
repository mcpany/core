#!/bin/bash
export PATH="$PATH:/usr/local/bin"
bazelisk test //ui:vitest_src_components_tools_rich_result_viewer_test_tsx
