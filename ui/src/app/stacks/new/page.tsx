/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { StackEditor } from "@/components/stacks/stack-editor";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

/**
 * NewStackPage component.
 * @returns The rendered component.
 */
export default function NewStackPage() {
  const router = useRouter();
  const { toast } = useToast();
  const [saving, setSaving] = useState(false);

  const handleSave = async (yamlContent: string, parsedJson: any) => {
      setSaving(true);
      try {
          // If name is not in YAML, defaults might be set by backend or we should validate
          if (!parsedJson.name) {
              throw new Error("Stack 'name' is required in YAML");
          }

          await apiClient.createCollection(parsedJson);
          toast({
              title: "Stack Created",
              description: `Stack "${parsedJson.name}" has been deployed.`
          });
          router.push("/stacks");
      } catch (e: any) {
          console.error("Failed to create stack", e);
          toast({
              variant: "destructive",
              title: "Deployment Failed",
              description: e.message || "Failed to create stack."
          });
      } finally {
          setSaving(false);
      }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6 h-[calc(100vh-4rem)] flex flex-col">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">New Stack</h1>
      </div>

      <Card className="flex-1 flex flex-col overflow-hidden">
          <CardHeader>
              <CardTitle>Define Stack</CardTitle>
              <CardDescription>
                  Enter your stack configuration in YAML format.
              </CardDescription>
          </CardHeader>
          <CardContent className="flex-1 overflow-hidden p-0">
              <div className="h-full p-6 pt-0">
                <StackEditor
                    onSave={handleSave}
                    onCancel={() => router.back()}
                    isSaving={saving}
                />
              </div>
          </CardContent>
      </Card>
    </div>
  );
}
