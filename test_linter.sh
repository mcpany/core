#!/bin/bash
cd server
go install honnef.co/go/tools/cmd/staticcheck@latest
~/go/bin/staticcheck ./pkg/tokenizer/...
