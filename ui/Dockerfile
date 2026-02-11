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
# Generate TypeScript Protobuf files inside Docker (Local Mode)
# -----------------------------------------------------------------------------
RUN apt-get update && apt-get install -y protobuf-compiler curl unzip findutils

# Download googleapis
WORKDIR /tmp
RUN curl -sSL https://github.com/googleapis/googleapis/archive/master.zip -o googleapis.zip && \
    unzip -q googleapis.zip && \
    mv googleapis-master /googleapis

# Create local proto directory in app
RUN mkdir -p /app/proto

# 1. Generate Repo Protos
# Input: /proto/... (relative to /proto)
# Output: /app/proto/...
WORKDIR /proto
RUN find . -name "*.proto" -exec protoc \
    --proto_path=. \
    --proto_path=/googleapis \
    --proto_path=/usr/include \
    --plugin=protoc-gen-ts_proto=/app/node_modules/.bin/protoc-gen-ts_proto \
    --ts_proto_out=/app/proto \
    --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
    {} +

# 2. Generate Standard Protos (google/protobuf/*.proto)
# Output: /app/proto/google/protobuf/...
WORKDIR /usr/include
RUN find google/protobuf -name "*.proto" -exec protoc \
    --proto_path=. \
    --plugin=protoc-gen-ts_proto=/app/node_modules/.bin/protoc-gen-ts_proto \
    --ts_proto_out=/app/proto \
    --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
    {} +

# 3. Generate Google APIs (google/api/*.proto)
# Output: /app/proto/google/api/...
WORKDIR /googleapis
RUN find google/api -name "*.proto" -exec protoc \
    --proto_path=. \
    --proto_path=/usr/include \
    --plugin=protoc-gen-ts_proto=/app/node_modules/.bin/protoc-gen-ts_proto \
    --ts_proto_out=/app/proto \
    --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
    {} +

WORKDIR /app

# Patch tsconfig.json to look for protos in ./proto instead of ../proto
# This ensures that TypeScript compilation finds the locally generated files
RUN sed -i 's|\.\./proto|./proto|g' tsconfig.json

# Verify tsconfig.json content
RUN cat tsconfig.json

# Verify generation (fail if missing)
# Note: we check /app/proto now
RUN echo "Checking for generated files..." && \
    ls -l /app/proto/config/v1/auth.ts || (echo "auth.ts MISSING" && exit 1) && \
    ls -l /app/proto/google/protobuf/struct.ts || (echo "struct.ts MISSING" && exit 1) && \
    ls -l /app/proto/google/api/annotations.ts || (echo "annotations.ts MISSING" && exit 1)

# -----------------------------------------------------------------------------

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
