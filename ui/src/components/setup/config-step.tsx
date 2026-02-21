/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { TemplateConfigForm } from "@/components/services/template-config-form";
import { ServiceTemplate } from "@/lib/templates";
import { Card, CardContent } from "@/components/ui/card";
import { Loader2 } from "lucide-react";

interface ConfigStepProps {
  template: ServiceTemplate;
  onSubmit: (values: Record<string, string>) => void;
  onCancel: () => void;
  loading: boolean;
}

export function ConfigStep({ template, onSubmit, onCancel, loading }: ConfigStepProps) {
  if (loading) {
      return (
          <div className="flex flex-col items-center justify-center space-y-4 min-h-[400px]">
              <Loader2 className="h-12 w-12 animate-spin text-primary" />
              <p className="text-muted-foreground animate-pulse">Configuring your service...</p>
          </div>
      );
  }

  return (
    <div className="w-full max-w-2xl">
      <div className="text-center space-y-2 mb-8">
            <h1 className="text-3xl font-bold tracking-tight">Configure Service</h1>
            <p className="text-muted-foreground text-lg">Enter the required details for {template.name}.</p>
      </div>
      <Card className="border-none shadow-lg bg-background/80 backdrop-blur-sm">
        <CardContent className="p-6">
          <TemplateConfigForm
              template={template}
              onSubmit={onSubmit}
              onCancel={onCancel}
          />
        </CardContent>
      </Card>
    </div>
  );
}
