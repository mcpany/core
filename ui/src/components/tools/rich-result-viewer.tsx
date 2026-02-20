/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { JsonView } from "@/components/ui/json-view";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card } from "@/components/ui/card";

interface RichResultViewerProps {
  result: any;
  className?: string;
}

export function RichResultViewer({ result, className }: RichResultViewerProps) {
  // Check for common output patterns
  let contentToRender = result;
  let defaultTab = "raw";

  // Case 1: Command Line Service Output (stdout)
  if (result && typeof result === 'object' && 'stdout' in result) {
      try {
          const parsed = JSON.parse(result.stdout);
          contentToRender = parsed;
          defaultTab = "parsed";
      } catch (e) {
          // Not JSON, keep raw stdout or full result?
          // If stdout is simple text, maybe just show it?
          // But full result has metadata like return_code.
          // Let's offer tabs.
      }
  }

  // Case 2: MCP Tool Result (content list)
  if (result && typeof result === 'object' && Array.isArray(result.content)) {
      // Extract text content
      const textContent = result.content
          .filter((c: any) => c.type === 'text')
          .map((c: any) => c.text)
          .join('\n');

      if (textContent) {
          try {
              const parsed = JSON.parse(textContent);
              contentToRender = parsed;
              defaultTab = "parsed";
          } catch (e) {
              // Not JSON
          }
      }
  }

  // Determine if we have a "primary" content distinct from the full raw result
  const hasPrimaryContent = contentToRender !== result;

  if (!hasPrimaryContent) {
      return (
          <Card className={className}>
              <JsonView data={result} smartTable={true} className="border-0 shadow-none" />
          </Card>
      );
  }

  return (
      <Card className={className}>
          <Tabs defaultValue="parsed" className="w-full">
              <div className="border-b px-4 py-2 bg-muted/20">
                  <TabsList className="h-8">
                      <TabsTrigger value="parsed" className="text-xs">Result</TabsTrigger>
                      <TabsTrigger value="full" className="text-xs">Full Output</TabsTrigger>
                  </TabsList>
              </div>

              <TabsContent value="parsed" className="p-0 m-0 border-0">
                  <JsonView data={contentToRender} smartTable={true} className="border-0 shadow-none rounded-none" />
              </TabsContent>

              <TabsContent value="full" className="p-0 m-0 border-0">
                  <JsonView data={result} smartTable={false} className="border-0 shadow-none rounded-none" />
              </TabsContent>
          </Tabs>
      </Card>
  );
}
