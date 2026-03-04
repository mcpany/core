# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

# Optimized GitHub Actions Runner Image for ARC
# Includes Go 1.24, Node 24, Playwright 1.58.0, and Go module caching.

FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive
USER root

# Create or rename the runner user to match the official actions-runner image (UID 1001)
RUN if id -u 1001 >/dev/null 2>&1; then \
  OLD_USER=$(id -un 1001); \
  if [ "$OLD_USER" != "runner" ]; then \
  groupmod -n runner $(id -gn 1001) && \
  usermod -l runner -d /home/runner -m $OLD_USER; \
  fi \
  else \
  groupadd -g 1001 runner && \
  useradd -m -u 1001 -g runner runner; \
  fi && \
  apt-get update && apt-get install -y sudo && \
  echo "runner ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Install base dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
  build-essential \
  ca-certificates \
  curl \
  findutils \
  git \
  grep \
  jq \
  make \
  sed \
  unzip \
  xz-utils \
  wget \
  python3 \
  python3-pip \
  && rm -rf /var/lib/apt/lists/*

# Install GitHub CLI (gh)
RUN mkdir -p -m 755 /etc/apt/keyrings \
  && wget -qO- https://cli.github.com/packages/githubcli-archive-keyring.gpg | tee /etc/apt/keyrings/githubcli-archive-keyring.gpg > /dev/null \
  && chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
  && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
  && apt-get update \
  && apt-get install -y gh \
  && rm -rf /var/lib/apt/lists/*

# Install Go 1.24 (Architecture-aware)
ENV GO_VERSION=1.24.0
RUN ARCH=$(dpkg --print-architecture | sed 's/x86_64/amd64/;s/aarch64/arm64/') && \
  wget -q https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz && \
  tar -C /usr/local -xzf go${GO_VERSION}.linux-${ARCH}.tar.gz && \
  rm go${GO_VERSION}.linux-${ARCH}.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

# Install Node.js 24
RUN curl -fsSL https://deb.nodesource.com/setup_24.x | bash - \
  && apt-get install -y nodejs \
  && rm -rf /var/lib/apt/lists/*

# Install Playwright 1.58.0 and browsers
RUN npm install -g playwright@1.58.0 \
  && npx playwright install --with-deps \
  && npm cache clean --force

# Pre-download Go modules
WORKDIR /tmp/warm-cache
COPY go.mod go.sum ./
COPY server/go.mod server/go.sum ./server/
# We need a dummy main.go or similar for go mod download to be most effective in some contexts,
# but go mod download works with just go.mod/go.sum.
RUN go mod download && cd server && go mod download \
  && rm -rf /tmp/warm-cache

# Final Cleanup
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Default to runner user for CI compatibility
USER runner
WORKDIR /home/runner/actions-runner
