/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { SecretsManager } from "@/components/settings/secrets-manager";

/**
 * Intent: Document SecretsPage
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
 * SecretsPage component.
 * @returns The rendered component.
 */
export default function SecretsPage() {
  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-8">
      <SecretsManager />
    </div>
  );
}
