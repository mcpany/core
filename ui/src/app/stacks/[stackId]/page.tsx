/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { use, useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { McpAnyManager } from "@/components/mcpany-manager";
import { StackEditor } from "@/components/stacks/stack-editor";
import { FileText, List, Box } from "lucide-react";

export default function StackDetailPage({ params }: { params: Promise<{ stackId: string }> }) {
  const { stackId } = use(params);

  return (
    <div className="space-y-6">
        <div className="flex flex-col gap-2">
            <h1 className="text-2xl font-bold tracking-tight flex items-center gap-2">
                <Box className="h-6 w-6 text-blue-500" />
                Stack: {stackId}
            </h1>
            <p className="text-muted-foreground">Manage services and configuration for this stack.</p>
        </div>

        <Tabs defaultValue="services" className="space-y-4">
            <TabsList>
                <TabsTrigger value="services" className="flex items-center gap-2">
                    <List className="h-4 w-4" /> Services
                </TabsTrigger>
                <TabsTrigger value="editor" className="flex items-center gap-2">
                    <FileText className="h-4 w-4" /> Editor
                </TabsTrigger>
            </TabsList>
            <TabsContent value="services" className="space-y-4">
                 {/* Reusing existing manager but maybe we want to wrap it or refactor it to fit better */}
                 <McpAnyManager />
            </TabsContent>
            <TabsContent value="editor">
                <StackEditor stackId={stackId} />
            </TabsContent>
        </Tabs>
    </div>
  );
}
