/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { NetworkGraphClient } from "@/components/network/network-graph-client";

/**
 * Intent: Document NetworkPage
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - Documented below.
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * NetworkPage component.
 * @returns The rendered component.
 */
export default function NetworkPage() {
  return (
    <div className="flex flex-col h-[calc(100vh-4rem)]">
      <NetworkGraphClient />
    </div>
  );
}
