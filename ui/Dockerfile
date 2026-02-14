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
ARG TARGETARCH
WORKDIR /app

# Install protoc to generate typescript definitions
RUN apt-get update && apt-get install -y curl unzip
RUN PROTOC_ARCH="x86_64"; \
    if [ "$TARGETARCH" = "arm64" ]; then PROTOC_ARCH="aarch_64"; fi; \
    curl -LO "https://github.com/protocolbuffers/protobuf/releases/download/v29.3/protoc-29.3-linux-$PROTOC_ARCH.zip" && \
    unzip protoc-29.3-linux-$PROTOC_ARCH.zip -d /usr/local && \
    rm protoc-29.3-linux-$PROTOC_ARCH.zip

COPY --from=deps /app/node_modules ./node_modules
COPY ui/ .
COPY --from=deps /app/package-lock.json ./package-lock.json
COPY proto/ /proto/
# Symlink node_modules so proto files can find dependencies
RUN ln -s /app/node_modules /proto/node_modules

# Generate TS protos (since they are not in the build context in CI)
RUN find /proto -name "*.proto" -exec protoc \
    --proto_path=/proto \
    --plugin=protoc-gen-ts_proto=./node_modules/.bin/protoc-gen-ts_proto \
    --ts_proto_out=/proto \
    --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
    {} +

# Fix relative imports in generated files to point to local google protos (if any generated)
# This matches the behavior in the root Makefile
RUN find /proto -name "*.ts" -exec sed -i 's|\.\./\.\./\.\./google|\.\./\.\./google|g' {} +

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
