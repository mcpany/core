/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { SecretManager } from "@/components/secrets/secret-manager";

export default function SecretsPage() {
  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-8">
      <SecretManager />
    </div>
  );
}
