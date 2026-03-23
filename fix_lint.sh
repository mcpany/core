#!/bin/bash
git checkout server/pkg/app/api.go
# Remove any handleStackConfig references safely
sed -i '/a\.handleStackConfig/d' server/pkg/app/api.go
# Make sure we didn't miss api_stacks_test
rm -f server/pkg/app/api_stacks_test.go
