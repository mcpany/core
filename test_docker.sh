#!/bin/bash
cat << 'INNER' > Dockerfile.test
FROM mirror.gcr.io/library/debian:bookworm-slim
RUN mkdir -p /app
WORKDIR /app
INNER
docker build -t test_docker -f Dockerfile.test .
