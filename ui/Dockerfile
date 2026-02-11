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
COPY proto/ /proto/
# Symlink node_modules so proto files can find dependencies
RUN ln -s /app/node_modules /proto/node_modules

# -----------------------------------------------------------------------------
# Generate TypeScript Protobuf files inside Docker
# -----------------------------------------------------------------------------
RUN apt-get update && apt-get install -y protobuf-compiler curl unzip findutils

# Download googleapis
WORKDIR /tmp
RUN curl -sSL https://github.com/googleapis/googleapis/archive/master.zip -o googleapis.zip && \
    unzip -q googleapis.zip && \
    mv googleapis-master /googleapis

# Change to root to simplify relative path handling for protoc
WORKDIR /

# Verify sources exist (fail if not)
RUN echo "Checking for source .proto files..." && \
    ls -l /proto/config/v1/auth.proto || (echo "Source auth.proto MISSING!" && exit 1)

# 1. Generate Repo Protos
# Use relative paths from root (e.g. proto/config/v1/auth.proto)
# Output to . (root), so proto/config/v1/auth.ts lands in /proto/config/v1/
RUN find proto -name "*.proto" -exec protoc \
    --proto_path=. \
    --proto_path=googleapis \
    --proto_path=usr/include \
    --plugin=protoc-gen-ts_proto=/app/node_modules/.bin/protoc-gen-ts_proto \
    --ts_proto_out=. \
    --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
    {} +

# 2. Generate Standard Protos (google/protobuf/*.proto)
# Output to proto/ (so they land in /proto/google/protobuf/...)
RUN find usr/include/google/protobuf -name "*.proto" -exec protoc \
    --proto_path=usr/include \
    --plugin=protoc-gen-ts_proto=/app/node_modules/.bin/protoc-gen-ts_proto \
    --ts_proto_out=proto \
    --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
    {} +

# 3. Generate Google APIs (google/api/*.proto)
# Output to proto/ (so they land in /proto/google/api/...)
RUN find googleapis/google/api -name "*.proto" -exec protoc \
    --proto_path=googleapis \
    --proto_path=usr/include \
    --plugin=protoc-gen-ts_proto=/app/node_modules/.bin/protoc-gen-ts_proto \
    --ts_proto_out=proto \
    --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
    {} +

# Verify generation (fail if missing)
RUN echo "Checking for generated files..." && \
    ls -l /proto/config/v1/auth.ts || (echo "auth.ts MISSING" && exit 1) && \
    ls -l /proto/google/protobuf/struct.ts || (echo "struct.ts MISSING" && exit 1) && \
    ls -l /proto/google/api/annotations.ts || (echo "annotations.ts MISSING" && exit 1)

# -----------------------------------------------------------------------------
# Restore WORKDIR for build
WORKDIR /app

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
