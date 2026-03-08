#!/bin/bash
find server -type f -name "*.go" -print0 | xargs -0 grep -Hn "invalid character"
