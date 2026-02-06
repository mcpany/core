/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { apiClient, DiscoveryProviderStatus } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Loader2, RefreshCcw, AlertTriangle } from "lucide-react";
import { formatDistanceToNow } from "date-fns";

export function DiscoveryDashboard() {
  const [statuses, setStatuses] = useState<DiscoveryProviderStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchStatus = async () => {
    setLoading(true);
    try {
      const res = await apiClient.getDiscoveryStatus();
      setStatuses(res.providers || []);
      setError(null);
    } catch (e: any) {
      console.error("Failed to fetch discovery status", e);
      setError(e.message || "Failed to load discovery status");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStatus();
    // Poll every 10 seconds
    const interval = setInterval(fetchStatus, 10000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium">Auto-Discovery Providers</h3>
          <p className="text-sm text-muted-foreground">
            Status of background service discovery agents.
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={fetchStatus} disabled={loading}>
          {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <RefreshCcw className="mr-2 h-4 w-4" />}
          Refresh
        </Button>
      </div>

      {error && (
        <div className="p-4 rounded-md bg-destructive/15 text-destructive text-sm flex items-center">
          <AlertTriangle className="h-4 w-4 mr-2" />
          {error}
        </div>
      )}

      {!loading && statuses.length === 0 && !error && (
        <div className="text-center p-8 text-muted-foreground border border-dashed rounded-lg">
          No discovery providers configured.
        </div>
      )}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {statuses.map((provider) => (
          <Card key={provider.name}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {provider.name}
              </CardTitle>
              {provider.status === "OK" ? (
                <Badge variant="outline" className="text-green-500 border-green-500/20 bg-green-500/10">
                  Active
                </Badge>
              ) : (
                <Badge variant="destructive">
                  {provider.status}
                </Badge>
              )}
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{provider.discoveredCount}</div>
              <p className="text-xs text-muted-foreground">
                Services Discovered
              </p>

              <div className="mt-4 text-xs space-y-2">
                <div className="flex justify-between">
                    <span className="text-muted-foreground">Last Run:</span>
                    <span className="font-mono">
                        {provider.lastRunAt ? formatDistanceToNow(new Date(provider.lastRunAt), { addSuffix: true }) : "Never"}
                    </span>
                </div>
                {provider.lastError && (
                    <div className="text-destructive mt-2 bg-destructive/10 p-2 rounded text-[10px] break-all">
                        {provider.lastError}
                    </div>
                )}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
