# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

FROM mirror.gcr.io/library/python:3.11-slim AS builder
WORKDIR /app
COPY main.py .
RUN pip install --no-cache-dir cowsay --target /app/deps

FROM mirror.gcr.io/library/python:3.11-slim
WORKDIR /app
COPY --from=builder /app /app
ENV PYTHONPATH=/app/deps
RUN useradd -m appuser
USER appuser
CMD ["python", "main.py"]
