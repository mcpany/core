/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { PlaygroundClient } from "@/components/playground/playground-client";

export default function PlaygroundPage() {
  return (
    <div className="flex flex-col h-[calc(100vh-5rem)]">
      <PlaygroundClient />
    </div>
  );
}
