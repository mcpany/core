/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { AuditLogViewer } from "@/components/audit/audit-log-viewer";

/**
 * AuditPage component.
 * Renders the audit logs page, which includes the AuditLogViewer.
 *
 * @returns The rendered AuditPage component.
 */
export default function AuditPage() {
    return (
        <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-3xl font-bold tracking-tight">Audit Logs</h2>
                    <p className="text-muted-foreground">Track and inspect tool executions for compliance and security.</p>
                </div>
            </div>
            <AuditLogViewer />
        </div>
    );
}
