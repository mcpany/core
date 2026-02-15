# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

FROM node:20-bookworm-slim AS base
# libc6-compat is not needed on Debian/Ubuntu base
# RUN apk add --no-cache libc6-compat

# Install dependencies only when needed
FROM base AS deps
WORKDIR /app

# Install dependencies based on the preferred package manager
COPY ui/package.json ui/package-lock.json* ui/.npmrc* ./
# Remove lockfile to force regeneration and avoid "Failed to patch lockfile" error on different architecture/OS
RUN rm -f package-lock.json && npm install -g npm@latest && npm install --no-audit --registry=https://registry.npmjs.org/

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY ui/ .
COPY --from=deps /app/package-lock.json ./package-lock.json
COPY proto/ /app/proto/
# Symlink node_modules so proto files can find dependencies
RUN ln -s /app/node_modules /app/proto/node_modules

# Install dependencies for protobuf generation
RUN apt-get update && apt-get install -y curl unzip

# Install protoc manually (apt version is too old for edition="2023")
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then PB_ARCH="x86_64"; \
    elif [ "$ARCH" = "aarch64" ]; then PB_ARCH="aarch_64"; \
    else echo "Unsupported arch: $ARCH"; exit 1; fi && \
    curl -sSL -o /tmp/protoc.zip "https://github.com/protocolbuffers/protobuf/releases/download/v29.3/protoc-29.3-linux-${PB_ARCH}.zip" && \
    unzip -q /tmp/protoc.zip -d /tmp/protoc && \
    mv /tmp/protoc/bin/protoc /usr/local/bin/ && \
    mv /tmp/protoc/include/* /usr/local/include/ && \
    rm -rf /tmp/protoc /tmp/protoc.zip

# Download external protos (googleapis and grpc-gateway) required for generation
RUN curl -sSL -o /tmp/googleapis.zip https://github.com/googleapis/googleapis/archive/refs/heads/master.zip && \
    unzip -q /tmp/googleapis.zip -d /tmp/ && \
    mv /tmp/googleapis-master /tmp/googleapis

RUN curl -sSL -o /tmp/grpc-gateway.zip https://github.com/grpc-ecosystem/grpc-gateway/archive/refs/tags/v2.27.3.zip && \
    unzip -q /tmp/grpc-gateway.zip -d /tmp/ && \
    mv /tmp/grpc-gateway-2.27.3 /tmp/grpc-gateway

# Generate TypeScript protobuf files
# We use the installed protoc and the ts-proto plugin from node_modules
# We set --proto_path to both /app/proto (to resolve source files) and /app (to resolve imports like "proto/config/...")
RUN find /app/proto -name "*.proto" -exec protoc \
    --proto_path=/app/proto \
    --proto_path=/app \
    --proto_path=/tmp/grpc-gateway \
    --proto_path=/tmp/googleapis \
    --plugin=protoc-gen-ts_proto=/app/node_modules/.bin/protoc-gen-ts_proto \
    --ts_proto_out=/app/proto \
    --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
    {} +

# Next.js collects completely anonymous telemetry data about general usage.
# Learn more here: https://nextjs.org/telemetry
# Uncomment the following line in case you want to disable telemetry during the build.
ENV NEXT_TELEMETRY_DISABLED=1

# Set default API URL for build time (prerendering) if not provided
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL:-http://localhost:50050}

RUN npm run build
RUN mkdir -p public

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# User authentication service URL
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL:-http://localhost:50050}

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Handle public directory if it exists
COPY --from=builder /app/public ./public


# Set the correct permission for prerender cache
RUN mkdir .next
RUN chown nextjs:nodejs .next

# Automatically leverage output traces to reduce image size
# https://nextjs.org/docs/advanced-features/output-file-tracing
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT=3000
# set hostname to localhost
ENV HOSTNAME=0.0.0.0

CMD ["node", "server.js"]
