/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient, UpstreamServiceConfig, ServiceDiagnosisResult } from "@/lib/client";
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { Stethoscope, CheckCircle2, AlertTriangle, XCircle, Loader2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";

export function ServiceDoctor() {
  const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
  const [selectedService, setSelectedService] = useState<string>("");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<ServiceDiagnosisResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    apiClient.listServices().then(setServices).catch(console.error);
  }, []);

  const runDiagnosis = async () => {
    if (!selectedService) return;
    setLoading(true);
    setResult(null);
    setError(null);
    try {
      const res = await apiClient.diagnoseService(selectedService);
      setResult(res);
    } catch (err) {
      console.error(err);
      setError("Failed to run diagnostics. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status.toLowerCase()) {
      case "ok":
      case "healthy":
        return <CheckCircle2 className="h-5 w-5 text-green-500" />;
      case "warning":
      case "degraded":
        return <AlertTriangle className="h-5 w-5 text-yellow-500" />;
      default:
        return <XCircle className="h-5 w-5 text-red-500" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
        case "ok":
        case "healthy":
          return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400 border-green-200 dark:border-green-800";
        case "warning":
        case "degraded":
          return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400 border-yellow-200 dark:border-yellow-800";
        default:
          return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400 border-red-200 dark:border-red-800";
      }
  };

  return (
    <Card className="h-full flex flex-col border-l-4 border-l-blue-500 bg-gradient-to-r from-background to-blue-50/5 dark:to-blue-900/10">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Stethoscope className="h-5 w-5 text-blue-500" />
          Service Doctor
        </CardTitle>
        <CardDescription>
          Diagnose connectivity and health issues for a specific service.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex-1 space-y-4">
        <div className="flex gap-2">
          <Select value={selectedService} onValueChange={setSelectedService}>
            <SelectTrigger className="flex-1">
              <SelectValue placeholder="Select a service to diagnose..." />
            </SelectTrigger>
            <SelectContent>
              {services.map((s) => (
                <SelectItem key={s.name} value={s.name}>
                  {s.name} <span className="text-muted-foreground text-xs">({s.httpService?.address || "N/A"})</span>
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button onClick={runDiagnosis} disabled={!selectedService || loading}>
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : "Diagnose"}
          </Button>
        </div>

        {error && (
          <Alert variant="destructive">
            <AlertTriangle className="h-4 w-4" />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {result && (
          <div className={`rounded-lg border p-4 ${getStatusColor(result.status)}`}>
            <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2 font-semibold text-lg">
                    {getStatusIcon(result.status)}
                    {result.status}
                </div>
                <Badge variant="outline" className="bg-background/50 backdrop-blur-sm">
                    {result.serviceName}
                </Badge>
            </div>
            <p className="text-sm font-medium mb-1">Message:</p>
            <p className="text-sm opacity-90 mb-2 whitespace-pre-wrap">{result.message}</p>

            {result.error && (
                <div className="mt-4 p-3 bg-red-500/10 rounded border border-red-500/20">
                    <p className="text-xs font-bold text-red-600 dark:text-red-400 uppercase tracking-wider mb-1">Detailed Error</p>
                    <code className="text-xs font-mono break-all text-red-700 dark:text-red-300">
                        {result.error}
                    </code>
                </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
