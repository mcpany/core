/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ShieldCheck, Clock, CheckCircle, XCircle } from "lucide-react";

/**
 * HITLPage component.
 * Displays the HITL Approval Interface with placeholders for pending approvals.
 * @returns The rendered component.
 */
export default function HITLPage() {
  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col overflow-hidden">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">HITL Approvals</h1>
          <p className="text-muted-foreground">Review and manage pending agent actions requiring human approval.</p>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Pending Approvals
            </CardTitle>
            <Clock className="h-4 w-4 text-amber-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">0</div>
            <p className="text-xs text-muted-foreground">
              Actions waiting for human review.
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Approved Actions
            </CardTitle>
            <CheckCircle className="h-4 w-4 text-emerald-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">0</div>
            <p className="text-xs text-muted-foreground">
              Actions successfully approved.
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Denied Actions
            </CardTitle>
            <XCircle className="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">0</div>
            <p className="text-xs text-muted-foreground">
              Actions denied execution.
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="flex-1 mt-4">
        <Card className="h-full flex flex-col">
            <CardHeader>
                <CardTitle>Approval Queue</CardTitle>
                <CardDescription>Actions currently awaiting review.</CardDescription>
            </CardHeader>
            <CardContent className="flex-1 flex items-center justify-center text-muted-foreground">
                <div className="flex flex-col items-center">
                    <ShieldCheck className="h-12 w-12 mb-4 opacity-50" />
                    <p>No pending approvals in the queue.</p>
                </div>
            </CardContent>
        </Card>
      </div>
    </div>
  );
}
