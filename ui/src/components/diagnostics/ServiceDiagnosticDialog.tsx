/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { apiClient, ServiceDiagnosticResult } from "@/lib/client";
import { Loader2, CheckCircle2, AlertTriangle, XCircle, Activity } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

interface ServiceDiagnosticDialogProps {
  serviceName: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ServiceDiagnosticDialog({
  serviceName,
  open,
  onOpenChange,
}: ServiceDiagnosticDialogProps) {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<ServiceDiagnosticResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open && serviceName) {
      runDiagnosis();
    } else {
      setResult(null);
      setError(null);
    }
  }, [open, serviceName]);

  const runDiagnosis = async () => {
    setLoading(true);
    setResult(null);
    setError(null);
    try {
      const res = await apiClient.diagnoseService(serviceName);
      setResult(res);
    } catch (err: any) {
      setError(err.message || "Failed to run diagnosis");
    } finally {
      setLoading(false);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "OK":
        return <CheckCircle2 className="h-10 w-10 text-green-500" />;
      case "WARNING":
        return <AlertTriangle className="h-10 w-10 text-yellow-500" />;
      case "ERROR":
        return <XCircle className="h-10 w-10 text-red-500" />;
      default:
        return <Activity className="h-10 w-10 text-muted-foreground" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "OK":
        return "text-green-600";
      case "WARNING":
        return "text-yellow-600";
      case "ERROR":
        return "text-red-600";
      default:
        return "text-muted-foreground";
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Service Diagnostic</DialogTitle>
          <DialogDescription>
            Running connectivity and configuration checks for <strong>{serviceName}</strong>.
          </DialogDescription>
        </DialogHeader>

        <div className="py-6">
          {loading ? (
            <div className="flex flex-col items-center justify-center space-y-4 text-center p-8">
              <Loader2 className="h-12 w-12 animate-spin text-primary" />
              <p className="text-muted-foreground">Diagnosing service health...</p>
            </div>
          ) : error ? (
            <Alert variant="destructive">
              <XCircle className="h-4 w-4" />
              <AlertTitle>Diagnostic Failed</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          ) : result ? (
            <div className="flex flex-col space-y-4">
              <div className="flex flex-col items-center justify-center p-6 border rounded-lg bg-muted/20">
                {getStatusIcon(result.status)}
                <h3 className={`mt-4 text-lg font-semibold ${getStatusColor(result.status)}`}>
                  {result.status}
                </h3>
                <p className="text-center mt-2 text-sm text-muted-foreground">
                  {result.message}
                </p>
              </div>

              {result.error && (
                <div className="bg-red-50 p-4 rounded-md border border-red-100 text-sm">
                  <p className="font-semibold text-red-800 mb-1">Error Details:</p>
                  <pre className="whitespace-pre-wrap text-red-700 font-mono text-xs">
                    {JSON.stringify(result.error, null, 2)}
                  </pre>
                </div>
              )}

              {result.status === "OK" && (
                <Alert>
                  <CheckCircle2 className="h-4 w-4 text-green-600" />
                  <AlertTitle>Everything looks good!</AlertTitle>
                  <AlertDescription>
                    The service is reachable and responding correctly.
                  </AlertDescription>
                </Alert>
              )}
            </div>
          ) : null}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
          {!loading && (
            <Button onClick={runDiagnosis}>Run Again</Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
