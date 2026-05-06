# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

FROM mirror.gcr.io/library/python:3.11-slim AS builder
WORKDIR /app
RUN pip install --no-cache-dir cowsay

FROM mirror.gcr.io/library/python:3.11-slim
WORKDIR /app
COPY --from=builder /usr/local/lib/python3.11/site-packages /usr/local/lib/python3.11/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin
COPY main.py .

RUN useradd -m -u 1000 -U -s /bin/sh -d /app appuser && \
    chown -R appuser:appuser /app
USER appuser

ENTRYPOINT ["python", "main.py"]
