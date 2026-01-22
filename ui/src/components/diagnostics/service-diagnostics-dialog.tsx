"use client";

import { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { CheckCircle2, XCircle, Loader2, SkipForward } from "lucide-react";
import { apiClient, DiagnosticReport } from "@/lib/client";
import { ScrollArea } from "@/components/ui/scroll-area";

interface ServiceDiagnosticsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  serviceName: string;
}

export function ServiceDiagnosticsDialog({ open, onOpenChange, serviceName }: ServiceDiagnosticsDialogProps) {
  const [report, setReport] = useState<DiagnosticReport | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open && serviceName) {
      runDiagnostics();
    } else {
        // Reset state when closed
        setReport(null);
        setError(null);
        setLoading(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, serviceName]);

  const runDiagnostics = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await apiClient.diagnoseService(serviceName);
      setReport(data);
    } catch (err: any) {
      setError(err.message || 'Failed to run diagnostics');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Service Diagnostics: {serviceName}</DialogTitle>
          <DialogDescription>
            Running health checks and connectivity tests.
          </DialogDescription>
        </DialogHeader>

        <div className="py-4">
          {loading && !report ? (
            <div className="flex flex-col items-center justify-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
              <p className="mt-2 text-sm text-muted-foreground">Running diagnostics...</p>
            </div>
          ) : error ? (
            <div className="rounded-md bg-destructive/10 p-4 text-destructive">
              <p className="font-semibold">Error running diagnostics</p>
              <p className="text-sm">{error}</p>
              <Button variant="outline" size="sm" onClick={runDiagnostics} className="mt-4">Retry</Button>
            </div>
          ) : report ? (
            <ScrollArea className="h-[400px] pr-4">
                <div className="space-y-4">
                    <div className="flex items-center justify-between border-b pb-2">
                        <span className="font-semibold">Overall Status</span>
                        <StatusBadge status={report.overall} />
                    </div>

                    <div className="space-y-4">
                        {report.steps.map((step, idx) => (
                            <div key={idx} className="rounded-lg border p-3">
                                <div className="flex items-center justify-between">
                                    <div className="flex items-center gap-2">
                                        <StatusIcon status={step.status} />
                                        <span className="font-medium">{step.name}</span>
                                    </div>
                                    <span className="text-xs text-muted-foreground">{step.duration_ms}ms</span>
                                </div>
                                {step.message && (
                                    <p className={`mt-2 text-sm ${step.status === 'failed' ? 'text-destructive font-medium' : 'text-muted-foreground'}`}>
                                        {step.message}
                                    </p>
                                )}
                                {step.details && (
                                    <pre className="mt-2 w-full overflow-x-auto rounded bg-muted p-2 text-xs font-mono">
                                        {step.details}
                                    </pre>
                                )}
                            </div>
                        ))}
                    </div>
                </div>
            </ScrollArea>
          ) : null}
        </div>
      </DialogContent>
    </Dialog>
  );
}

function StatusBadge({ status }: { status: string }) {
    const styles: Record<string, string> = {
        pending: "bg-muted text-muted-foreground",
        running: "bg-blue-100 text-blue-700",
        success: "bg-green-100 text-green-700",
        failed: "bg-red-100 text-red-700",
        skipped: "bg-yellow-100 text-yellow-700",
    };
    return (
        <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium capitalize ${styles[status] || styles.pending}`}>
            {status}
        </span>
    );
}

function StatusIcon({ status }: { status: string }) {
    switch (status) {
        case 'success': return <CheckCircle2 className="h-5 w-5 text-green-500" />;
        case 'failed': return <XCircle className="h-5 w-5 text-destructive" />;
        case 'running': return <Loader2 className="h-5 w-5 animate-spin text-blue-500" />;
        case 'skipped': return <SkipForward className="h-5 w-5 text-yellow-500" />;
        default: return <div className="h-5 w-5 rounded-full border-2 border-muted" />;
    }
}
