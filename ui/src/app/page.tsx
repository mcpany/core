/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { McpAnyManager } from "@/components/mcpany-manager";

/**
 * The home page component.
 *
 * @returns The main dashboard page.
 */
export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center bg-background p-4 sm:p-8">
      <McpAnyManager />
    </main>
  );
}
