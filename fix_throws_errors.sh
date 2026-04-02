#!/bin/bash
find src server -name "*.go" | xargs sed -i 's/\/\/ Throws\/Errors:/\/\/ Errors:/g'
