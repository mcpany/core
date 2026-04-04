/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Network } from "lucide-react";

/**
 * Summary: UI for managing and auto-discovering MCP servers across transports.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.JSX.Element: The rendered UnifiedDiscoveryManager component.
 *
 * Throws/Errors:
 *   - None.
 */
export function UnifiedDiscoveryManager() {
  const transports = [
    { id: "tx-1", name: "Stdio Transport", status: "Connected" },
    { id: "tx-2", name: "SSE Transport", status: "Disconnected" },
    { id: "tx-3", name: "WebSocket Transport", status: "Connected" },
  ];

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div className="space-y-1">
          <CardTitle className="text-sm font-medium">Unified Discovery Manager</CardTitle>
          <CardDescription>
            UI for managing and auto-discovering MCP servers across transports.
          </CardDescription>
        </div>
        <Network className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="space-y-4 mt-4">
          {transports.map((transport) => (
            <div key={transport.id} className="flex items-center justify-between border-b pb-2 last:border-0 last:pb-0">
              <div>
                <p className="text-sm font-medium">{transport.name}</p>
                <p className="text-xs text-muted-foreground">{transport.id}</p>
              </div>
              <div className="text-right">
                <span className={`text-xs font-medium px-2 py-1 rounded-full ${transport.status === "Connected" ? "bg-green-100 text-green-700" : "bg-red-100 text-red-700"}`}>
                  {transport.status}
                </span>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
