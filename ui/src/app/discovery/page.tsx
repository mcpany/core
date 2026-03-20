/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { DiscoveryManager } from "@/components/discovery/discovery-manager";

/**
 * DiscoveryPage component.
 * @returns The rendered component.
 */
export default function DiscoveryPage() {
  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-4 md:p-8 space-y-4 overflow-y-auto">
      <DiscoveryManager />
    </div>
  );
}
