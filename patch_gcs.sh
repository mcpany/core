#!/bin/bash

# Update gcs.go to check for fs or fs.client nilness to avoid panic
sed -i 's/rc, err := f.fs.client.Bucket(f.fs.bucket).Object(f.name).NewRangeReader/if f.fs == nil || f.fs.client == nil { return 0, fmt.Errorf("file not opened for reading") }\n\trc, err := f.fs.client.Bucket(f.fs.bucket).Object(f.name).NewRangeReader/' server/pkg/upstream/filesystem/provider/gcs.go
