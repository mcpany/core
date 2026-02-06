# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

FROM node:20-bookworm-slim AS base
# libc6-compat is not needed on Debian/Ubuntu base
# RUN apk add --no-cache libc6-compat

# Install dependencies only when needed
FROM base AS deps
# Avoid WORKDIR to prevent overlayfs mount issues in CI
# Just use default root dir and copy explicitly

# Install dependencies based on the preferred package manager
COPY ui/package.json ui/package-lock.json* ui/.npmrc* /app/
# Remove lockfile to force regeneration and avoid "Failed to patch lockfile" error on different architecture/OS
RUN cd /app && rm -f package-lock.json && npm install -g npm@latest && npm install --no-audit --registry=https://registry.npmjs.org/

# Rebuild the source code only when needed
FROM base AS builder
# Manually create app dir without WORKDIR
# RUN mkdir -p /app - No need if we copy to it directly, but explicit is fine
COPY --from=deps /app/node_modules /app/node_modules
COPY ui/ /app/
COPY --from=deps /app/package-lock.json /app/package-lock.json
COPY proto/ /proto/
# Symlink node_modules so proto files can find dependencies
RUN ln -s /app/node_modules /proto/node_modules

# Install tools to generate protos inside the container
RUN apt-get update && apt-get install -y protobuf-compiler curl unzip

# Download googleapis
RUN curl -sSL -o googleapis.zip https://github.com/googleapis/googleapis/archive/refs/heads/master.zip && \
    unzip -q googleapis.zip && \
    mv googleapis-master /googleapis && \
    rm googleapis.zip

# Generate local protos
RUN find /proto -name "*.proto" -exec protoc \
    --proto_path=/proto \
    --proto_path=/googleapis \
    --plugin=protoc-gen-ts_proto=/app/node_modules/.bin/protoc-gen-ts_proto \
    --ts_proto_out=/proto \
    --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
    {} +

# Generate standard protos (google/api, google/protobuf)
RUN find /usr/include/google/protobuf /googleapis/google/api -name "*.proto" -exec protoc \
    --proto_path=/usr/include \
    --proto_path=/googleapis \
    --plugin=protoc-gen-ts_proto=/app/node_modules/.bin/protoc-gen-ts_proto \
    --ts_proto_out=/proto \
    --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
    {} +

# Next.js collects completely anonymous telemetry data about general usage.
# Learn more here: https://nextjs.org/telemetry
# Uncomment the following line in case you want to disable telemetry during the build.
ENV NEXT_TELEMETRY_DISABLED=1

# Set default API URL for build time (prerendering) if not provided
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL:-http://localhost:50050}

RUN cd /app && npm run build
RUN mkdir -p /app/public

# Production image, copy all the files and run next
FROM base AS runner
# RUN mkdir -p /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# User authentication service URL
ENV NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL:-http://localhost:50050}

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Handle public directory if it exists
COPY --from=builder /app/public /app/public


# Set the correct permission for prerender cache
RUN mkdir -p /app/.next
RUN chown nextjs:nodejs /app/.next

# Automatically leverage output traces to reduce image size
# https://nextjs.org/docs/advanced-features/output-file-tracing
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone /app/
COPY --from=builder --chown=nextjs:nodejs /app/.next/static /app/.next/static

USER nextjs

EXPOSE 3000

ENV PORT=3000
# set hostname to localhost
ENV HOSTNAME=0.0.0.0

CMD ["sh", "-c", "cd /app && node server.js"]
