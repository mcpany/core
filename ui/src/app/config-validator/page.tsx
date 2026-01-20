/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import Editor from "@monaco-editor/react";
import { Button } from "@/components/ui/button";
import { Loader2, CheckCircle2, XCircle } from "lucide-react";
import { useTheme } from "next-themes";
import { toast } from "sonner";

export default function ConfigValidatorPage() {
  const [content, setContent] = useState("");
  const [isValidating, setIsValidating] = useState(false);
  const [result, setResult] = useState<{ valid: boolean; errors?: string[] } | null>(
    null
  );
  const { theme } = useTheme();

  const handleValidate = async () => {
    if (!content.trim()) {
      toast.error("Please enter some configuration to validate.");
      return;
    }

    setIsValidating(true);
    setResult(null);

    try {
      const response = await fetch("/api/v1/config/validate", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          // Add auth headers if needed, assuming cookie/session handled or injected globally
        },
        body: JSON.stringify({ content }),
      });

      if (!response.ok) {
        throw new Error(`Server returned ${response.status}`);
      }

      const data = await response.json();
      setResult(data);

      if (data.valid) {
        toast.success("Configuration is valid!");
      } else {
        toast.error("Configuration has errors.");
      }
    } catch (error) {
      console.error("Validation failed", error);
      toast.error("Failed to connect to validation service.");
      setResult({ valid: false, errors: [(error as Error).message] });
    } finally {
      setIsValidating(false);
    }
  };

  return (
    <div className="h-full flex flex-col p-6 space-y-4">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Config Validator</h1>
          <p className="text-muted-foreground">
            Validate your YAML or JSON configuration against the server schema.
          </p>
        </div>
        <Button onClick={handleValidate} disabled={isValidating}>
          {isValidating && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Validate Configuration
        </Button>
      </div>

      <div className="flex-1 grid grid-cols-1 md:grid-cols-2 gap-6 min-h-0">
        <div className="flex flex-col space-y-2 h-full min-h-0">
          <div className="text-sm font-medium">Input (YAML/JSON)</div>
          <div className="flex-1 border rounded-md overflow-hidden min-h-[400px]">
            <Editor
              height="100%"
              defaultLanguage="yaml"
              theme={theme === "dark" ? "vs-dark" : "light"}
              value={content}
              onChange={(value) => setContent(value || "")}
              options={{
                minimap: { enabled: false },
                fontSize: 14,
                scrollBeyondLastLine: false,
              }}
            />
          </div>
        </div>

        <div className="flex flex-col space-y-2 h-full min-h-0">
          <div className="text-sm font-medium">Validation Results</div>
          <div className="flex-1 border rounded-md bg-muted/30 p-4 overflow-auto min-h-[400px]">
            {!result ? (
              <div className="h-full flex items-center justify-center text-muted-foreground">
                Enter configuration and click Validate to see results.
              </div>
            ) : result.valid ? (
              <div className="flex flex-col items-center justify-center h-full text-green-600 space-y-2">
                <CheckCircle2 className="h-16 w-16" />
                <span className="text-lg font-semibold">Valid Configuration</span>
                <p className="text-sm text-muted-foreground text-center max-w-xs">
                  The configuration syntax and structure match the server schema.
                </p>
              </div>
            ) : (
              <div className="space-y-4">
                <div className="flex items-center text-red-600 space-x-2">
                  <XCircle className="h-6 w-6" />
                  <span className="text-lg font-semibold">Validation Errors</span>
                </div>
                <div className="space-y-2">
                  {result.errors?.map((err, i) => (
                    <div
                      key={i}
                      className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-900 rounded text-sm font-mono text-red-800 dark:text-red-200 break-words whitespace-pre-wrap"
                    >
                      {err}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
