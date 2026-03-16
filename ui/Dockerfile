# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

# ---- deps stage: install all npm packages ----
FROM mirror.gcr.io/library/node:24-bookworm-slim AS deps
WORKDIR /app
COPY ui/package.json ui/package-lock.json* ui/.npmrc* ./
# Regenerate lockfile to avoid architecture/OS mismatch issues
RUN rm -f package-lock.json && \
    npm install --no-audit --registry=https://registry.npmjs.org/

# ---- builder stage: compile Vite SPA ----
FROM mirror.gcr.io/library/node:24-bookworm-slim AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY ui/ .
COPY proto/ /proto/
# Symlink node_modules so proto files can find dependencies during build
RUN ln -s /app/node_modules /proto/node_modules
RUN npm run build

# ---- runner stage: serve via vite preview (handles proxy to backend) ----
# vite preview reads vite.config.ts at runtime, so BACKEND_URL is injected
# as a runtime env var rather than baked into the bundle.
FROM gcr.io/distroless/nodejs24-debian12:nonroot AS runner
WORKDIR /app

# Copy only what is needed to run `vite preview`:
#   node_modules  - vite binary + plugins (including proxy config deps)
#   dist/         - the pre-built static SPA
#   vite.config.ts - proxy rules (read by vite preview at startup)
#   index.html + package.json - required by the Vite runtime
COPY --from=deps    --chown=65532:65532 /app/node_modules   ./node_modules
COPY --from=builder --chown=65532:65532 /app/dist           ./dist
COPY --from=builder --chown=65532:65532 /app/vite.config.ts ./vite.config.ts
COPY --from=builder --chown=65532:65532 /app/index.html     ./index.html
COPY --from=builder --chown=65532:65532 /app/package.json   ./package.json

EXPOSE 3000

# BACKEND_URL is the runtime address of the mcpany backend.
# Override this env var when deploying (e.g. via Helm values or docker --env).
ENV BACKEND_URL=http://localhost:50050
ENV PORT=3000

# `vite preview` serves the pre-built dist/ and proxies API paths to BACKEND_URL
# as configured in the preview.proxy section of vite.config.ts.
CMD ["node_modules/.bin/vite", "preview", "--host", "0.0.0.0", "--port", "3000"]
