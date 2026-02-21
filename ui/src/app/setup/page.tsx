/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { ArrowRight, Sparkles, CheckCircle2, Database, Terminal, FileText, Globe } from "lucide-react";
import { ServiceTemplateSelector } from "@/components/services/service-template-selector";
import { TemplateConfigForm } from "@/components/services/template-config-form";
import { ServiceTemplate } from "@/lib/templates";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { useRouter } from "next/navigation";
import { cn } from "@/lib/utils";

enum WizardStep {
  WELCOME = "WELCOME",
  TEMPLATE_SELECT = "TEMPLATE_SELECT",
  CONFIGURE = "CONFIGURE",
  SUCCESS = "SUCCESS",
}

export default function SetupPage() {
  const [step, setStep] = useState<WizardStep>(WizardStep.WELCOME);
  const [selectedTemplate, setSelectedTemplate] = useState<ServiceTemplate | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { toast } = useToast();
  const router = useRouter();

  const handleTemplateSelect = (template: ServiceTemplate) => {
    setSelectedTemplate(template);
    setStep(WizardStep.CONFIGURE);
  };

  const handleConfigSubmit = async (values: Record<string, string>) => {
    if (!selectedTemplate) return;
    setIsSubmitting(true);

    try {
      // 1. Prepare configuration based on template
      // Use 'config' property as per ServiceTemplate definition in lib/templates.ts
      // (Note: apiClient.getServiceTemplates uses 'serviceConfig' but local templates use 'config')
      const config = JSON.parse(JSON.stringify(selectedTemplate.config || (selectedTemplate as any).serviceConfig));

      // Apply values
      let configStr = JSON.stringify(config);
      Object.entries(values).forEach(([key, value]) => {
          // Replace ${KEY} safely
          // We construct regex for ${KEY} and replace it with the value string
          configStr = configStr.replace(new RegExp(`\\$\\{${key}\\}`, 'g'), () => value);
      });
      const finalConfig = JSON.parse(configStr);

      // Ensure basic fields
      finalConfig.id = ""; // New service
      finalConfig.disable = false;

      // 2. Register Service
      await apiClient.registerService(finalConfig);

      toast({
          title: "Service Connected",
          description: `Successfully configured ${selectedTemplate.name}.`
      });

      setStep(WizardStep.SUCCESS);

    } catch (e) {
      console.error("Failed to setup service", e);
      toast({
          variant: "destructive",
          title: "Setup Failed",
          description: e instanceof Error ? e.message : "An error occurred."
      });
    } finally {
        setIsSubmitting(false);
    }
  };

  const renderWelcome = () => (
    <div className="w-full max-w-lg animate-in fade-in slide-in-from-bottom-4 duration-500">
      <Card className="border-none shadow-2xl bg-background/80 backdrop-blur-xl">
        <CardHeader className="text-center pb-2">
          <div className="mx-auto bg-primary/10 p-4 rounded-full mb-4 w-16 h-16 flex items-center justify-center">
             <Sparkles className="w-8 h-8 text-primary" />
          </div>
          <CardTitle className="text-3xl font-bold">Welcome to MCP Any</CardTitle>
          <CardDescription className="text-lg mt-2">
            Your universal gateway for AI tools. Let's get your first service connected in seconds.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6 pt-6">
           <div className="grid grid-cols-2 gap-4">
               <div className="flex flex-col items-center gap-2 p-4 rounded-lg bg-muted/50 text-center">
                   <Globe className="w-6 h-6 text-blue-500" />
                   <span className="font-medium text-sm">Connect APIs</span>
               </div>
               <div className="flex flex-col items-center gap-2 p-4 rounded-lg bg-muted/50 text-center">
                   <Database className="w-6 h-6 text-green-500" />
                   <span className="font-medium text-sm">Query Databases</span>
               </div>
               <div className="flex flex-col items-center gap-2 p-4 rounded-lg bg-muted/50 text-center">
                   <Terminal className="w-6 h-6 text-orange-500" />
                   <span className="font-medium text-sm">Run Commands</span>
               </div>
               <div className="flex flex-col items-center gap-2 p-4 rounded-lg bg-muted/50 text-center">
                   <FileText className="w-6 h-6 text-purple-500" />
                   <span className="font-medium text-sm">Read Files</span>
               </div>
           </div>
        </CardContent>
        <CardFooter>
          <Button
            size="lg"
            className="w-full text-lg h-12 gap-2"
            onClick={() => setStep(WizardStep.TEMPLATE_SELECT)}
          >
            Get Started <ArrowRight className="w-5 h-5" />
          </Button>
        </CardFooter>
      </Card>
    </div>
  );

  const renderTemplateSelect = () => (
      <div className="w-full max-w-4xl space-y-4 animate-in fade-in slide-in-from-right-4 duration-500">
          <div className="flex items-center justify-between">
            <h2 className="text-2xl font-bold">Choose a Starter Template</h2>
            <Button variant="ghost" onClick={() => setStep(WizardStep.WELCOME)}>Back</Button>
          </div>
          <Card className="border-none shadow-lg bg-background/80 backdrop-blur-sm">
              <CardContent className="p-6">
                  <ServiceTemplateSelector onSelect={handleTemplateSelect} />
              </CardContent>
          </Card>
      </div>
  );

  const renderConfigure = () => (
      <div className="w-full max-w-lg animate-in fade-in slide-in-from-right-4 duration-500">
          <Card className="border-none shadow-xl bg-background/95 backdrop-blur">
              <CardContent className="p-6">
                  {selectedTemplate && (
                      <TemplateConfigForm
                        template={selectedTemplate}
                        onCancel={() => setStep(WizardStep.TEMPLATE_SELECT)}
                        onSubmit={handleConfigSubmit}
                      />
                  )}
              </CardContent>
          </Card>
      </div>
  );

  const renderSuccess = () => (
      <div className="w-full max-w-md text-center animate-in zoom-in-95 duration-500">
          <Card className="border-none shadow-2xl bg-background/80 backdrop-blur-xl">
              <CardContent className="pt-10 pb-10 space-y-6">
                  <div className="mx-auto bg-green-100 dark:bg-green-900/30 p-4 rounded-full w-20 h-20 flex items-center justify-center">
                      <CheckCircle2 className="w-10 h-10 text-green-600 dark:text-green-400" />
                  </div>
                  <div className="space-y-2">
                      <h2 className="text-3xl font-bold">You're All Set!</h2>
                      <p className="text-muted-foreground text-lg">
                          Your service <strong>{selectedTemplate?.name}</strong> is now active and ready to use.
                      </p>
                  </div>
                  <div className="flex flex-col gap-3 pt-4">
                      <Button size="lg" className="w-full" onClick={() => router.push('/')}>
                          Go to Dashboard
                      </Button>
                      <Button variant="outline" className="w-full" onClick={() => {
                          setSelectedTemplate(null);
                          setStep(WizardStep.TEMPLATE_SELECT);
                      }}>
                          Connect Another Service
                      </Button>
                  </div>
              </CardContent>
          </Card>
      </div>
  );

  return (
    <>
      {step === WizardStep.WELCOME && renderWelcome()}
      {step === WizardStep.TEMPLATE_SELECT && renderTemplateSelect()}
      {step === WizardStep.CONFIGURE && renderConfigure()}
      {step === WizardStep.SUCCESS && renderSuccess()}
    </>
  );
}
