/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { CheckCircle2, AlertCircle, Loader2 } from "lucide-react";
import { apiClient } from "@/lib/client";

export default function SchemaPlaygroundPage() {
  const [content, setContent] = useState("");
  const [isValidating, setIsValidating] = useState(false);
  const [result, setResult] = useState<{ valid: boolean; message: string } | null>(null);

  const EXAMPLES = {
    http: `upstream_services:
  - name: my-http-service
    http_service:
      address: http://localhost:8080
    authentication:
      api_key:
        header_name: X-API-Key
        api_key:
          plain_text: my-secret-token`,
    command: `upstream_services:
  - name: my-tool
    command_line_service:
      command: python3
      args: ["-m", "http.server", "8000"]
      working_dir: /tmp`,
    invalid: `upstream_services:
  - name: broken-service
    # Missing service type (e.g., http_service)`
  };

  const handleValidate = async () => {
    if (!content.trim()) return;

    setIsValidating(true);
    setResult(null);

    try {
      const response = await fetch("/api/v1/validate", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ content }),
      });

      const data = await response.json();
      setResult({
        valid: data.valid,
        message: data.message || data.error || "Unknown response",
      });
    } catch (err: any) {
      setResult({
        valid: false,
        message: "Failed to connect to validation API: " + err.message,
      });
    } finally {
      setIsValidating(false);
    }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Schema Validation Playground</h2>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
        <Card className="col-span-4">
          <CardHeader>
            <CardTitle>Configuration Editor</CardTitle>
            <CardDescription>
              Paste your JSON or YAML configuration snippet here to validate it against the server schema.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => setContent(EXAMPLES.http)}>Load HTTP Example</Button>
              <Button variant="outline" size="sm" onClick={() => setContent(EXAMPLES.command)}>Load Stdio Example</Button>
              <Button variant="outline" size="sm" onClick={() => setContent(EXAMPLES.invalid)}>Load Invalid Example</Button>
            </div>
            <Textarea
              placeholder="Paste JSON or YAML here..."
              className="min-h-[400px] font-mono"
              value={content}
              onChange={(e) => setContent(e.target.value)}
            />
            <div className="flex justify-end">
              <Button onClick={handleValidate} disabled={isValidating || !content.trim()}>
                {isValidating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Validate Configuration
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card className="col-span-3">
          <CardHeader>
            <CardTitle>Validation Results</CardTitle>
            <CardDescription>
              Real-time feedback on your configuration structure and rules.
            </CardDescription>
          </CardHeader>
          <CardContent>
            {result ? (
              <Alert variant={result.valid ? "default" : "destructive"}>
                {result.valid ? (
                  <CheckCircle2 className="h-4 w-4 text-green-600" />
                ) : (
                  <AlertCircle className="h-4 w-4" />
                )}
                <AlertTitle>{result.valid ? "Valid Configuration" : "Validation Error"}</AlertTitle>
                <AlertDescription className="mt-2 whitespace-pre-wrap">
                  {result.message}
                </AlertDescription>
              </Alert>
            ) : (
              <div className="flex flex-col items-center justify-center h-[200px] text-muted-foreground">
                <p>No validation results yet.</p>
                <p className="text-sm">Click "Validate Configuration" to start.</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
