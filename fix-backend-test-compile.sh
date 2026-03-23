#!/bin/bash
find server/pkg -name "*_test.go" | xargs sed -i 's/mock.Anything/mock.Anything()/g'
find server/pkg -name "*_test.go" | xargs sed -i 's/mock.Anything()()/mock.Anything()/g'
