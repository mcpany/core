/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { Card, CardHeader, CardContent } from "@/components/ui/card";

/**
 * WidgetSkeleton component.
 * Renders a placeholder skeleton for dashboard widgets while they are loading.
 * @returns The rendered component.
 */
export function WidgetSkeleton() {
  return (
    <Card className="h-full w-full animate-pulse border-muted/50 bg-background/50">
      <CardHeader className="flex flex-row items-center gap-4 space-y-0 p-6 pb-2">
        <div className="h-5 w-5 rounded-full bg-muted" />
        <div className="h-4 w-32 rounded bg-muted" />
      </CardHeader>
      <CardContent className="p-6 pt-4 space-y-3">
        <div className="h-20 w-full rounded bg-muted/50" />
        <div className="h-4 w-2/3 rounded bg-muted/30" />
      </CardContent>
    </Card>
  );
}
