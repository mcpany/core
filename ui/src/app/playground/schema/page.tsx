/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { CheckCircle2, AlertCircle, Loader2, Play } from "lucide-react";
import { apiClient } from "@/lib/client";
import { ConfigEditor } from "@/components/stacks/config-editor";

/**
 * SchemaPlaygroundPage provides an interactive environment for users to test and validate
 * configuration snippets against the MCP Any schema.
 * @returns The Schema Playground page component.
 */
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
      const data = await apiClient.validateConfig(content);
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
    <div className="flex-1 flex flex-col h-[calc(100vh-4rem)] p-8 pt-6 space-y-4">
      <div className="flex items-center justify-between shrink-0">
        <div>
            <h2 className="text-3xl font-bold tracking-tight">Schema Validator</h2>
            <p className="text-muted-foreground">Validate your configuration syntax and rules.</p>
        </div>
        <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => setContent(EXAMPLES.http)}>Load HTTP Example</Button>
            <Button variant="outline" size="sm" onClick={() => setContent(EXAMPLES.command)}>Load CLI Example</Button>
            <Button variant="outline" size="sm" onClick={() => setContent(EXAMPLES.invalid)}>Load Invalid Example</Button>
        </div>
      </div>

      <div className="flex-1 grid gap-4 md:grid-cols-2 lg:grid-cols-7 min-h-0">
        <Card className="col-span-4 flex flex-col h-full overflow-hidden">
          <CardHeader className="py-3 border-b bg-muted/20 flex flex-row items-center justify-between space-y-0">
            <div className="flex flex-col">
                <CardTitle className="text-sm font-medium">Editor</CardTitle>
            </div>
            <Button onClick={handleValidate} disabled={isValidating || !content.trim()} size="sm">
                {isValidating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Play className="mr-2 h-4 w-4" />}
                Validate
            </Button>
          </CardHeader>
          <CardContent className="flex-1 p-0 relative">
            <ConfigEditor
                value={content}
                onChange={(v) => setContent(v || "")}
                language="yaml"
            />
          </CardContent>
        </Card>

        <Card className="col-span-3 flex flex-col h-full overflow-hidden bg-muted/5">
          <CardHeader className="py-3 border-b bg-muted/20">
            <CardTitle className="text-sm font-medium">Results</CardTitle>
          </CardHeader>
          <CardContent className="p-4 flex-1 overflow-auto">
            {result ? (
              <Alert variant={result.valid ? "default" : "destructive"} className={result.valid ? "border-green-500 bg-green-50 dark:bg-green-900/20" : ""}>
                {result.valid ? (
                  <CheckCircle2 className="h-4 w-4 text-green-600 dark:text-green-400" />
                ) : (
                  <AlertCircle className="h-4 w-4" />
                )}
                <AlertTitle>{result.valid ? "Configuration Valid" : "Validation Failed"}</AlertTitle>
                <AlertDescription className="mt-2 font-mono text-xs whitespace-pre-wrap break-all">
                  {result.message}
                </AlertDescription>
              </Alert>
            ) : (
              <div className="flex flex-col items-center justify-center h-full text-muted-foreground opacity-50">
                <CheckCircle2 className="h-12 w-12 mb-2" />
                <p>Ready to validate.</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
