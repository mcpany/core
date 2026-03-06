#!/bin/bash
cat << 'INNEREOF' > Dockerfile.test
FROM alpine:latest
RUN echo "hello" > /hello.txt
INNEREOF
docker build -t test-alpine -f Dockerfile.test .
