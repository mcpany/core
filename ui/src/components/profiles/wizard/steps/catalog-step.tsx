/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { WizardService } from "../wizard-dialog";
import { Loader2 } from "lucide-react";

interface CatalogStepProps {
  onNext: (services: WizardService[]) => void;
}

/**
 * CatalogStep allows users to browse and select service templates.
 *
 * @param props - The component props.
 * @param props.onNext - Callback when the user proceeds to the next step.
 */
export function CatalogStep({ onNext }: CatalogStepProps) {
  const [templates, setTemplates] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedTemplateIds, setSelectedTemplateIds] = useState<Set<string>>(new Set());

  useEffect(() => {
    apiClient.getServiceTemplates()
      .then(setTemplates)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const toggleSelection = (id: string) => {
    const next = new Set(selectedTemplateIds);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    setSelectedTemplateIds(next);
  };

  const handleNext = () => {
    // Convert selected templates to initial WizardService objects
    const services: WizardService[] = Array.from(selectedTemplateIds).map(id => {
      const tmpl = templates.find(t => t.id === id) as any;
      return {
        templateId: id,
        instanceName: "", // To be filled in next step
        config: tmpl.serviceConfig,
        isAuthenticated: false
      };
    });
    onNext(services);
  };

  if (loading) {
    return <div className="flex justify-center p-12"><Loader2 className="animate-spin" /></div>;
  }

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {templates.map(tmpl => {
          const isSelected = selectedTemplateIds.has(tmpl.id);
          return (
            <Card
              key={tmpl.id}
              className={`cursor-pointer transition-all border-2 ${isSelected ? "border-primary bg-primary/5" : "border-transparent hover:border-muted-foreground/25"}`}
              onClick={() => toggleSelection(tmpl.id)}
            >
              <CardHeader>
                <div className="flex justify-between items-start">
                  <div className="flex gap-2 items-center">
                    {/* Icon could be added here */}
                    <CardTitle className="text-lg">{tmpl.name}</CardTitle>
                  </div>
                  <Checkbox checked={isSelected} />
                </div>
                <div className="flex gap-1 mt-2">
                  {tmpl.tags?.map((tag: string) => (
                    <Badge key={tag} variant="secondary" className="text-[10px]">{tag}</Badge>
                  ))}
                </div>
              </CardHeader>
              <CardContent>
                <CardDescription>{tmpl.description}</CardDescription>
              </CardContent>
            </Card>
          );
        })}
      </div>

      <div className="flex justify-end">
        <Button onClick={handleNext} disabled={selectedTemplateIds.size === 0}>
          Next: Configure ({selectedTemplateIds.size})
        </Button>
      </div>
    </div>
  );
}
