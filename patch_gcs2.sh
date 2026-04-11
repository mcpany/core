#!/bin/bash

# Update gcs_test.go to fix panic when calling Stat
sed -i 's/assert.Panics(t, func() { f.Stat() })/if _, err := f.Stat(); err == nil { t.Errorf("expected error from Stat()") }/' server/pkg/upstream/filesystem/provider/gcs_test.go
