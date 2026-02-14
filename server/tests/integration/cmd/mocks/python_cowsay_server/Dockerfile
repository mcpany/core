# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

FROM python:3.11-slim
WORKDIR /app
# Install build tools just in case, though cowsay is likely pure python
RUN apt-get update && apt-get install -y gcc build-essential && rm -rf /var/lib/apt/lists/*
COPY main.py .
RUN pip install --no-cache-dir cowsay
