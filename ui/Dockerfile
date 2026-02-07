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
# Copy proto folder into ui (instead of sibling) to ensure Next.js can resolve it within build context
COPY proto/ /app/proto/

# Install protoc manually (apt version is too old for 'edition' syntax)
ARG TARGETARCH
RUN apt-get update && apt-get install -y curl unzip && \
    PB_ARCH=$(if [ "$TARGETARCH" = "arm64" ]; then echo "aarch_64"; else echo "x86_64"; fi) && \
    curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v33.1/protoc-33.1-linux-${PB_ARCH}.zip && \
    unzip -q protoc-33.1-linux-${PB_ARCH}.zip -d /usr/local && \
    rm protoc-33.1-linux-${PB_ARCH}.zip && \
    mkdir -p /app/proto/google/api && \
    curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto -o /app/proto/google/api/annotations.proto && \
    curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto -o /app/proto/google/api/http.proto && \
    find /usr/local/include/google/protobuf -name "*.proto" -exec protoc --plugin=protoc-gen-ts_proto=./node_modules/.bin/protoc-gen-ts_proto --ts_proto_out=/app/proto --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web --proto_path=/usr/local/include {} + && \
    find /app/proto/google/api -name "*.proto" -exec protoc --plugin=protoc-gen-ts_proto=./node_modules/.bin/protoc-gen-ts_proto --ts_proto_out=/app/proto --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web --proto_path=/app/proto {} + && \
    find /app/proto -name "*.proto" -not -path "/app/proto/google/*" -type f -exec sed -i '/google\/protobuf\/go_features.proto/d' {} + && \
    find /app/proto -name "*.proto" -not -path "/app/proto/google/*" -type f -exec protoc --plugin=protoc-gen-ts_proto=./node_modules/.bin/protoc-gen-ts_proto --ts_proto_out=/app/proto --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web --proto_path=/app/proto --proto_path=/usr/local/include {} +

# Update tsconfig to use local proto path
RUN sed -i 's|\.\./proto|proto|g' tsconfig.json
# Symlink node_modules so proto files can find dependencies (if they import from node_modules)
RUN ln -s /app/node_modules /app/proto/node_modules

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
