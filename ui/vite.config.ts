/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";
import fs from "fs";

// Resolve the proto directory: in a Bazel build it is copied next to the
// project root; locally it lives one level up.
const localProto = path.join(__dirname, "proto");
const rootProto = path.join(__dirname, "../proto");
const protoPath = fs.existsSync(localProto) ? localProto : rootProto;

// Resolve the @bufbuild/protobuf/wire sub-path (needed by generated proto files)
const bufbuildWirePath = path.join(
  __dirname,
  "node_modules/.pnpm/@bufbuild+protobuf@2.11.0/node_modules/@bufbuild/protobuf/dist/esm/wire/index.js"
);
const bufbuildWire = fs.existsSync(bufbuildWirePath)
  ? bufbuildWirePath
  : path.join(__dirname, "node_modules/@bufbuild/protobuf/dist/esm/wire/index.js");

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.join(__dirname, "src"),
      "@proto": protoPath,
      "@google": path.join(protoPath, "google"),
      "@bufbuild/protobuf/wire": bufbuildWire,
      // grpc-web and browser-headers use CommonJS builds; point Vite at the
      // UMD browser bundle so Rollup can resolve them properly.
      "@improbable-eng/grpc-web": path.join(
        __dirname,
        "node_modules/@improbable-eng/grpc-web/dist/grpc-web-client.umd.js"
      ),
      "browser-headers": path.join(
        __dirname,
        "node_modules/browser-headers/dist/browser-headers.umd.js"
      ),
      // 'long' (required by ts-proto generated code) points to the UMD build
      "long": path.join(__dirname, "node_modules/long/umd/index.js"),
    },
  },
  server: {
    port: 9002,
    // In development, proxy API and gRPC-Web calls to the backend.
    proxy: {
      "/api/v1": {
        target: process.env.BACKEND_URL || "http://localhost:50050",
        changeOrigin: true,
      },
      // gRPC-Web: service name segments start with "mcpany."
      "/mcpany.": {
        target: process.env.BACKEND_URL || "http://localhost:50050",
        changeOrigin: true,
      },
      "/doctor": {
        target: process.env.BACKEND_URL || "http://localhost:50050",
        changeOrigin: true,
      },
      "/v1/": {
        target: process.env.BACKEND_URL || "http://localhost:50050",
        changeOrigin: true,
      },
      "/auth/oauth/": {
        target: process.env.BACKEND_URL || "http://localhost:50050",
        changeOrigin: true,
      },
      "/auth/login": {
        target: process.env.BACKEND_URL || "http://localhost:50050",
        changeOrigin: true,
      },
      "/debug/": {
        target: process.env.BACKEND_URL || "http://localhost:50050",
        changeOrigin: true,
      },
      "/sse": {
        target: process.env.BACKEND_URL || "http://localhost:50050",
        changeOrigin: true,
      },
      "/messages": {
        target: process.env.BACKEND_URL || "http://localhost:50050",
        changeOrigin: true,
      },
      "/mcp/": {
        target: process.env.BACKEND_URL || "http://localhost:50050",
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: "dist",
    sourcemap: true,
  },
  css: {
    devSourcemap: true,
  },
});
