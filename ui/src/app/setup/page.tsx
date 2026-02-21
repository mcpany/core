/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { ServiceTemplate } from "@/lib/templates";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { applyTemplateFields } from "@/lib/template-utils";

// Components will be implemented in the next step
import { WelcomeStep } from "@/components/setup/welcome-step";
import { TemplateSelectorStep } from "@/components/setup/template-selector-step";
import { ConfigStep } from "@/components/setup/config-step";
import { SuccessStep } from "@/components/setup/success-step";

type Step = "WELCOME" | "TEMPLATE" | "CONFIG" | "SUCCESS";

/**
 * SetupPage orchestrates the First Run Wizard flow.
 */
export default function SetupPage() {
  const [step, setStep] = useState<Step>("WELCOME");
  const [selectedTemplate, setSelectedTemplate] = useState<ServiceTemplate | null>(null);
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  const handleStart = () => setStep("TEMPLATE");

  const handleTemplateSelect = (template: ServiceTemplate) => {
    setSelectedTemplate(template);
    setStep("CONFIG");
  };

  const handleConfigSubmit = async (values: Record<string, string>) => {
    if (!selectedTemplate) return;
    setLoading(true);
    try {
      const configuredConfig = applyTemplateFields(selectedTemplate, values);

      // Ensure defaults and cleanup
      const newService = {
          ...configuredConfig,
          id: "", // ensure new ID generation on backend if not provided
          version: configuredConfig.version || "1.0.0",
          priority: configuredConfig.priority || 0,
          disable: false,
      };

      await apiClient.registerService(newService);
      setStep("SUCCESS");
    } catch (error) {
      console.error("Setup failed", error);
      toast({
        title: "Setup Failed",
        description: error instanceof Error ? error.message : "Could not register service.",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleCancelConfig = () => {
    setSelectedTemplate(null);
    setStep("TEMPLATE");
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-[calc(100vh-100px)] p-4 max-w-4xl mx-auto w-full">
      {step === "WELCOME" && <WelcomeStep onStart={handleStart} />}
      {step === "TEMPLATE" && <TemplateSelectorStep onSelect={handleTemplateSelect} />}
      {step === "CONFIG" && selectedTemplate && (
        <ConfigStep
          template={selectedTemplate}
          onSubmit={handleConfigSubmit}
          onCancel={handleCancelConfig}
          loading={loading}
        />
      )}
      {step === "SUCCESS" && <SuccessStep />}
    </div>
  );
}
