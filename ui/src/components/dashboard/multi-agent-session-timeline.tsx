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
import { Clock } from "lucide-react";

/**
 * Summary: Displays a visual tracking of agent handoffs and shared tool state.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - React.JSX.Element: The rendered MultiAgentSessionTimeline component.
 *
 * Throws/Errors:
 *   - None.
 */
export function MultiAgentSessionTimeline() {
  const sessions = [
    { id: "sess-1", status: "Active", time: "10:00 AM", details: "Handoff to Agent A" },
    { id: "sess-2", status: "Completed", time: "09:45 AM", details: "Shared Tool State Updated" },
    { id: "sess-3", status: "Pending", time: "09:30 AM", details: "Waiting for Agent B" },
  ];

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <div className="space-y-1">
          <CardTitle className="text-sm font-medium">Multi-Agent Session Timeline</CardTitle>
          <CardDescription>
            Visual tracking of agent handoffs and shared tool state.
          </CardDescription>
        </div>
        <Clock className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="space-y-4 mt-4">
          {sessions.map((session) => (
            <div key={session.id} className="flex items-center justify-between border-b pb-2 last:border-0 last:pb-0">
              <div>
                <p className="text-sm font-medium">{session.id}</p>
                <p className="text-xs text-muted-foreground">{session.details}</p>
              </div>
              <div className="text-right">
                <p className="text-sm font-medium">{session.time}</p>
                <p className="text-xs text-muted-foreground">{session.status}</p>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
