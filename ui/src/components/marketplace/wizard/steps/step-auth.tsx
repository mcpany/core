/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useWizard } from "../wizard-context";
import { useEffect, useState } from "react";
import { apiClient, Credential } from "@/lib/client";
import { Loader2, ShieldCheck, AlertTriangle } from "lucide-react";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

export function StepAuth() {
  const { state, updateConfig } = useWizard();
  const { config } = state;
  const [credentials, setCredentials] = useState<Credential[]>([]);
  const [loading, setLoading] = useState(true);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{success: boolean, message: string} | null>(null);

  useEffect(() => {
    const fetchCredentials = async () => {
      try {
        const creds = await apiClient.listCredentials();
        setCredentials(creds);
      } catch (e) {
        console.error("Failed to load credentials", e);
      } finally {
        setLoading(false);
      }
    };
    fetchCredentials();
  }, []);

  const handleAuthChange = (value: string) => {
      const cred = credentials.find(c => c.id === value);
      if (cred) {
          // Placeholder for handling auth change logic
      }
  };

  const [selectedCredId, setSelectedCredId] = useState<string>("");

  const getServiceType = (c: any) => {
      if (c.httpService) return "http";
      if (c.grpcService) return "grpc";
      if (c.commandLineService) return "command_line";
      if (c.mcpService) return "mcp";
      if (c.serviceType) return c.serviceType; // Fallback if manually set in state
      return "unknown";
  };

  const testConnection = async () => {
      setTesting(true);
      setTestResult(null);
      try {
          const res = await apiClient.testAuth({
              credential_id: selectedCredId,
              service_type: getServiceType(config),
              service_config: config
          });
          setTestResult({
              success: res.success,
              message: res.message || (res.success ? "Connection successful" : "Connection failed")
          });
      } catch (e) {
          console.error("Auth test error:", e);
          setTestResult({ success: false, message: "Connection failed (Network/Server Error)" });
      } finally {
          setTesting(false);
      }
  };

  if (loading) {
      return <div className="flex justify-center p-8"><Loader2 className="h-6 w-6 animate-spin" /></div>;
  }

  return (
    <div className="space-y-6">
      <Alert variant="default" className="bg-amber-50 dark:bg-amber-900/20 border-amber-200 dark:border-amber-800">
        <AlertTriangle className="h-4 w-4 text-amber-600 dark:text-amber-400" />
        <AlertTitle className="text-amber-800 dark:text-amber-300">Test Connection Only</AlertTitle>
        <AlertDescription className="text-amber-700 dark:text-amber-400">
          Authentication selected here is used <strong>only to test connectivity</strong> during this configuration.
          The actual credential binding for the running service must be configured in the <strong>Upstream Services</strong> page after instantiation.
        </AlertDescription>
      </Alert>

      <div className="space-y-4">
          <Label>Select Credential for Testing</Label>
          <Select onValueChange={(val) => { setSelectedCredId(val); handleAuthChange(val); }}>
              <SelectTrigger>
                  <SelectValue placeholder="Select a credential..." />
              </SelectTrigger>
              <SelectContent>
                  {credentials.map(c => (
                      <SelectItem key={c.id} value={c.id}>{c.name}</SelectItem>
                  ))}
                  {credentials.length === 0 && (
                       <SelectItem value="none" disabled>No credentials found</SelectItem>
                  )}
              </SelectContent>
          </Select>

          <div className="pt-4">
              <Button
                onClick={testConnection}
                disabled={!selectedCredId || testing}
                variant="secondary"
                className="w-full sm:w-auto"
              >
                  {testing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <ShieldCheck className="mr-2 h-4 w-4" />}
                  Test Connection
              </Button>
          </div>

          {testResult && (
              <Alert variant={testResult.success ? "default" : "destructive"} className={testResult.success ? "border-green-200 bg-green-50 dark:bg-green-900/10" : ""}>
                   <AlertDescription className={testResult.success ? "text-green-800 dark:text-green-300" : ""}>
                       {testResult.message}
                   </AlertDescription>
              </Alert>
          )}
      </div>
    </div>
  );
}
