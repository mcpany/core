#!/bin/bash
sed -i 's/5\*time.Second, 100\*time.Millisecond, "server should be ready to accept connections"/5*time.Second, 10*time.Millisecond, "server should be ready to accept connections"/g' server/pkg/app/server_test.go
sed -i 's/5\*time.Second, 100\*time.Millisecond)/5*time.Second, 10*time.Millisecond)/g' server/pkg/app/server_test.go
