/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useState } from "react";
import { ServiceTemplate } from "@/lib/templates";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { UpstreamServiceConfig, apiClient, ToolDefinition } from "@/lib/client";
import { applyTemplateFields } from "@/lib/template-utils";
import { Loader2, CheckCircle2, AlertCircle } from "lucide-react";
import { DiscoveredToolsViewer } from "@/components/services/editor/discovered-tools-viewer";

interface SmartConnectionWizardProps {
  template: ServiceTemplate;
  onCancel: () => void;
  onComplete: (config: UpstreamServiceConfig) => void;
}

export function SmartConnectionWizard({ template, onCancel, onComplete }: SmartConnectionWizardProps) {
  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [values, setValues] = useState<Record<string, string>>({});
  const [serviceName, setServiceName] = useState<string>(template.config?.name || template.name.toLowerCase().replace(/\s+/g, '-'));
  const [isValidating, setIsValidating] = useState(false);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [discoveredTools, setDiscoveredTools] = useState<ToolDefinition[]>([]);
  const [configuredConfig, setConfiguredConfig] = useState<UpstreamServiceConfig | null>(null);

  const handleChange = (name: string, value: string) => {
    setValues((prev) => ({ ...prev, [name]: value }));
  };

  const handleConnect = async (e: React.FormEvent) => {
    e.preventDefault();
    setStep(2);
    setIsValidating(true);
    setValidationError(null);

    try {
      // 1. Apply fields
      const configPartial = applyTemplateFields(template, values);

      // 2. Override name
      configPartial.name = serviceName;
      configPartial.id = ""; // Ensure it's treated as new
      configPartial.version = configPartial.version || "1.0.0";
      configPartial.priority = configPartial.priority || 0;
      configPartial.disable = false;

      const configToValidate = configPartial as UpstreamServiceConfig;

      // 3. Validate
      const response = await apiClient.validateService(configToValidate);

      if (response && response.valid === false) {
          setValidationError(response.error || "Validation failed.");
          setIsValidating(false);
          return;
      }

      // 4. Capture tools
      if (response && response.discoveredTools) {
          setDiscoveredTools(response.discoveredTools);
      } else {
          setDiscoveredTools([]);
      }

      setConfiguredConfig(configToValidate);
      setIsValidating(false);
      setStep(3);
    } catch (err: unknown) {
      setValidationError(err instanceof Error ? err.message : "An unexpected error occurred during connection testing.");
      setIsValidating(false);
    }
  };

  const handleFinish = () => {
      if (configuredConfig) {
          onComplete(configuredConfig);
      }
  };

  const Icon = template.icon;

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-300">
      <CardHeader className="px-0">
        <div className="flex items-center gap-2">
            <div className="p-2 bg-primary/10 rounded-md">
                {typeof Icon === "string" ? null : Icon ? <Icon className="w-6 h-6 text-primary" /> : null}
            </div>
            <div>
                <CardTitle>{template.name}</CardTitle>
                <CardDescription>{template.description}</CardDescription>
            </div>
        </div>
      </CardHeader>

      {step === 1 && (
        <form onSubmit={handleConnect} className="space-y-6">
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="serviceName">Service Name</Label>
              <Input
                id="serviceName"
                placeholder="my-service"
                value={serviceName}
                onChange={(e) => setServiceName(e.target.value)}
                required
              />
            </div>

            {template.fields?.map((field) => (
              <div key={field.name} className="space-y-2">
                <Label htmlFor={field.name}>{field.label}</Label>
                <Input
                  id={field.name}
                  type={field.type || "text"}
                  placeholder={field.placeholder}
                  value={values[field.name] || ""}
                  onChange={(e) => handleChange(field.name, e.target.value)}
                  required
                />
                {field.replaceToken && (
                    <p className="text-xs text-muted-foreground">
                        This will replace <code>{field.replaceToken}</code> in the configuration.
                    </p>
                )}
              </div>
            ))}
          </div>

          <div className="flex justify-end gap-2 mt-8">
            <Button type="button" variant="outline" onClick={onCancel}>
              Cancel
            </Button>
            <Button type="submit">
              Connect
            </Button>
          </div>
        </form>
      )}

      {step === 2 && (
          <div className="py-12 flex flex-col items-center justify-center space-y-4">
              {isValidating ? (
                  <>
                      <Loader2 className="h-10 w-10 text-primary animate-spin" />
                      <div className="text-center">
                          <h3 className="text-lg font-medium">Testing Connection...</h3>
                          <p className="text-sm text-muted-foreground">Validating configuration and discovering tools.</p>
                      </div>
                  </>
              ) : validationError ? (
                  <>
                      <div className="p-3 bg-destructive/10 rounded-full">
                          <AlertCircle className="h-10 w-10 text-destructive" />
                      </div>
                      <div className="text-center space-y-2 max-w-md">
                          <h3 className="text-lg font-medium text-destructive">Connection Failed</h3>
                          <p className="text-sm text-muted-foreground bg-muted p-2 rounded break-all whitespace-pre-wrap text-left">
                              {validationError}
                          </p>
                      </div>
                      <div className="flex gap-2 mt-4">
                          <Button variant="outline" onClick={() => setStep(1)}>
                              Back to Edit
                          </Button>
                      </div>
                  </>
              ) : null}
          </div>
      )}

      {step === 3 && (
          <div className="space-y-6">
              <div className="flex items-center gap-3 p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg">
                  <CheckCircle2 className="h-8 w-8 text-green-600 dark:text-green-400" />
                  <div>
                      <h3 className="font-medium text-green-900 dark:text-green-100">Connection Successful</h3>
                      <p className="text-sm text-green-700 dark:text-green-300">
                          Successfully connected to the upstream service.
                      </p>
                  </div>
              </div>

              <div className="max-h-[300px] overflow-y-auto pr-2">
                  <DiscoveredToolsViewer tools={discoveredTools} />
              </div>

              <div className="flex justify-end gap-2 mt-8 pt-4 border-t">
                  <Button variant="outline" onClick={() => setStep(1)}>
                      Back
                  </Button>
                  <Button onClick={handleFinish}>
                      Save & Finish
                  </Button>
              </div>
          </div>
      )}
    </div>
  );
}
