/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { AttestationDashboard } from "@/components/audit/attestation-dashboard";

export default function AttestationPage() {
  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] p-4 md:p-8 space-y-6">
      <div className="flex flex-col gap-2 shrink-0">
        <h1 className="text-3xl font-bold tracking-tight">Supply Chain Attestation</h1>
        <p className="text-muted-foreground">
          Security dashboard for verifying the provenance and cryptographic signatures of connected MCP servers.
        </p>
      </div>
      <div className="flex-1 min-h-0">
        <AttestationDashboard />
      </div>
    </div>
  );
}
