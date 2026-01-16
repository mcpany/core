/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useToast } from "@/hooks/use-toast";
import { apiClient } from "@/lib/client";
import { UpstreamServiceConfig } from "@/lib/types";
import { Credential } from "@proto/config/v1/auth";
import { Plus, RotateCw, ChevronLeft, ArrowLeft } from "lucide-react";
import { SERVICE_TEMPLATES, ServiceTemplate } from "@/lib/templates";
import { ServiceConfigDiff } from "./services/service-config-diff";
import { cn } from "@/lib/utils";

const formSchema = z.object({
  name: z.string().min(1, "Name is required"),
  type: z.enum(["grpc", "http", "command_line", "openapi", "other"]),
  address: z.string().optional(),
  command: z.string().optional(),
  configJson: z.string().optional(), // For advanced mode
  upstreamAuth: z.any().optional(), // Store auth config object
  tags: z.string().optional(),
});

interface RegisterServiceDialogProps {
  onSuccess?: () => void;
  trigger?: React.ReactNode;
  serviceToEdit?: UpstreamServiceConfig;
}

export function RegisterServiceDialog({ onSuccess, trigger, serviceToEdit }: RegisterServiceDialogProps) {
  const [open, setOpen] = useState(false);
  const [credentials, setCredentials] = useState<Credential[]>([]);
  const [view, setView] = useState<"templates" | "form">("templates");
  const [selectedTemplate, setSelectedTemplate] = useState<ServiceTemplate | null>(null);
  const [showDiff, setShowDiff] = useState(false);
  const [pendingConfig, setPendingConfig] = useState<UpstreamServiceConfig | null>(null);

  const { toast } = useToast();
  const isEditing = !!serviceToEdit;

  // Reset view when dialog opens/closes
  const handleOpenChange = (newOpen: boolean) => {
      setOpen(newOpen);
      if (!newOpen) {
          setView("templates");
          setSelectedTemplate(null);
          setShowDiff(false);
          setPendingConfig(null);
          form.reset();
      } else if (isEditing) {
          setView("form");
      }
  };

  const defaultValues = serviceToEdit ? {
      name: serviceToEdit.name,
      type: (serviceToEdit.grpcService ? "grpc" :
            serviceToEdit.httpService ? "http" :
            serviceToEdit.commandLineService ? "command_line" :
            serviceToEdit.openapiService ? "openapi" : "other") as "grpc" | "http" | "command_line" | "openapi" | "other",
      address: serviceToEdit.grpcService?.address || serviceToEdit.httpService?.address || serviceToEdit.openapiService?.address || "",
      command: serviceToEdit.commandLineService?.command || "",
      configJson: JSON.stringify(serviceToEdit, null, 2),
      upstreamAuth: serviceToEdit.upstreamAuth,
      tags: serviceToEdit.tags?.join(", ") || "",
  } : {
      name: "",
      type: "http" as const,
      address: "",
      command: "",
      configJson: "{\n  \"name\": \"my-service\",\n  \"httpService\": {\n    \"address\": \"https://api.example.com\"\n  }\n}",
      upstreamAuth: undefined,
      tags: "",
  };

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues,
  });

  const loadCredentials = async () => {
      try {
          const data = await apiClient.listCredentials();
          setCredentials(data);
      } catch (e) {
          console.error("Failed to load credentials", e);
      }
  };

  const handleCredentialSelect = (credId: string) => {
      const cred = credentials.find(c => c.id === credId);
      if (cred && cred.authentication) {
          form.setValue("upstreamAuth", cred.authentication, { shouldDirty: true });
          toast({ title: "Authentication Applied", description: `Applied auth from '${cred.name}'` });
      }
  };

  // Load credentials when dialog opens or tab changes to auth
  if (open && credentials.length === 0) {
      loadCredentials();
  }

  const handleTemplateSelect = (template: ServiceTemplate) => {
      setSelectedTemplate(template);

      // Pre-fill form based on template config
      const config = template.config;
      const type = config.grpcService ? "grpc" :
                   config.httpService ? "http" :
                   config.commandLineService ? "command_line" :
                   config.openapiService ? "openapi" : "other";

      form.setValue("name", config.name || "");
      form.setValue("type", type as any);

      if (config.httpService?.address) form.setValue("address", config.httpService.address);
      if (config.grpcService?.address) form.setValue("address", config.grpcService.address);
      if (config.openapiService?.address) form.setValue("address", config.openapiService.address);
      if (config.commandLineService?.command) form.setValue("command", config.commandLineService.command);
      if (config.tags) form.setValue("tags", config.tags.join(", "));

      // Also set the JSON for advanced usage
      form.setValue("configJson", JSON.stringify(config, null, 2));

      setView("form");
  };

  const performSave = async (config: UpstreamServiceConfig) => {
    try {
        if (isEditing) {
            await apiClient.updateService(config);
            toast({ title: "Service Updated", description: `${config.name} updated successfully.` });
        } else {
            await apiClient.registerService(config);
            toast({ title: "Service Registered", description: `${config.name} registered successfully.` });
        }

        setOpen(false);
        if (onSuccess) onSuccess();
    } catch (error: any) {
        console.error(error);
        toast({
            variant: "destructive",
            title: isEditing ? "Update Failed" : "Registration Failed",
            description: error.message || "Something went wrong.",
        });
    }
  };

  const onSubmit = async (values: z.infer<typeof formSchema>) => {
    try {
      let config: UpstreamServiceConfig;

      if (values.type === 'other') {
          if (!values.configJson) throw new Error("Config JSON is required for 'Other' type");
          config = JSON.parse(values.configJson);
      } else {
          // Construct config
          config = {
              name: values.name,
              id: serviceToEdit?.id || "",
              version: serviceToEdit?.version || "1.0.0",
              disable: serviceToEdit?.disable ?? false,
              priority: serviceToEdit?.priority ?? 0,
              loadBalancingStrategy: serviceToEdit?.loadBalancingStrategy ?? 0,
              sanitizedName: serviceToEdit?.sanitizedName || "",
              callPolicies: serviceToEdit?.callPolicies || [],
              preCallHooks: serviceToEdit?.preCallHooks || [],
              postCallHooks: serviceToEdit?.postCallHooks || [],
              prompts: serviceToEdit?.prompts || [],
              autoDiscoverTool: serviceToEdit?.autoDiscoverTool ?? false,
              configError: "",
              tags: values.tags ? values.tags.split(",").map(t => t.trim()).filter(t => t) : [],
          };

          if (values.type === 'grpc') {
              config.grpcService = { address: values.address || "", useReflection: true, tools: [], resources: [], calls: {}, prompts: [], protoCollection: [], protoDefinitions: [], tlsConfig: undefined, healthCheck: undefined };
          } else if (values.type === 'http') {
              config.httpService = { address: values.address || "", tools: [], calls: {}, resources: [], prompts: [], healthCheck: undefined, tlsConfig: undefined };
          } else if (values.type === 'command_line') {
              config.commandLineService = { command: values.command || "", workingDirectory: "", local: false, env: {}, tools: [], resources: [], prompts: [], communicationProtocol: 0, calls: {}, healthCheck: undefined, cache: undefined, containerEnvironment: undefined, timeout: undefined };

              // Try to preserve environment variables from configJson if available.
              // This allows users to use the "Advanced (JSON)" tab to set env vars (like keys)
              // while still using the "Basic" tab for command editing.
              if (values.configJson) {
                  try {
                      const jsonConfig = JSON.parse(values.configJson);
                      if (jsonConfig.commandLineService?.env) {
                          config.commandLineService.env = jsonConfig.commandLineService.env;
                      }
                  } catch (e) {
                      // Ignore JSON parse errors here
                  }
              }

          } else if (values.type === 'openapi') {
               config.openapiService = { address: values.address || "", tools: [], calls: {}, resources: [], prompts: [], specContent: undefined, specUrl: undefined, healthCheck: undefined, tlsConfig: undefined };
          }

          if (values.upstreamAuth) {
              config.upstreamAuth = values.upstreamAuth;
          }
      }

      // Diff Logic
      if (isEditing && !showDiff) {
          // Basic comparison - in production we might want a more robust equality check
          // that ignores order or specific noisy fields.
          // For now, strict JSON equality is a safe start.

          // To reduce noise, we can temporarily mask fields that we know API adds but we don't control,
          // OR we assume serviceToEdit and config are close enough in structure.

          // Since config is reconstructed from form, it might miss some deep fields if they were in serviceToEdit but not in form.
          // The form reconstruction logic above tries to preserve top-level fields (callPolicies, etc).
          // However, if the user was using "Advanced JSON" mode, they might have edited anything.

          // Let's proceed with showing diff if ANY change is detected or if we just want to confirm.
          // Actually, we should probably ALWAYS show diff on edit for safety if it's a "Premium" feature?
          // No, only if changed is better UX.

          // Note: JSON.stringify order might differ. But usually consistently produced if created same way.
          // serviceToEdit comes from Go JSON marshaling. config comes from JS object.
          // Keys might be different order.
          // But `ServiceConfigDiff` uses `yaml.dump({ sortKeys: true })` so it handles visualization well.

          // Let's set pending and show diff.
          setPendingConfig(config);
          setShowDiff(true);
          return;
      }

      await performSave(config);

    } catch (error: any) {
      console.error(error);
      toast({
        variant: "destructive",
        title: isEditing ? "Update Failed" : "Registration Failed",
        description: error.message || "Something went wrong.",
      });
    }
  };

  const selectedType = form.watch("type");

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        {trigger || <Button><Plus className="mr-2 h-4 w-4" /> Register Service</Button>}
      </DialogTrigger>
      <DialogContent className={cn("max-h-[85vh] overflow-y-auto transition-all duration-200", showDiff ? "sm:max-w-[900px]" : "sm:max-w-[700px]")}>
        <DialogHeader>
          <DialogTitle>
              {showDiff ? (
                  <div className="flex items-center gap-2">
                      <Button variant="ghost" size="icon" className="-ml-2 h-6 w-6" onClick={() => setShowDiff(false)}>
                          <ArrowLeft className="h-4 w-4" />
                      </Button>
                      Review Changes
                  </div>
              ) : view === "form" && !isEditing ? (
                  <div className="flex items-center gap-2">
                    <Button variant="ghost" size="icon" className="-ml-2 h-6 w-6" onClick={() => setView("templates")} aria-label="Back to templates">
                        <ChevronLeft className="h-4 w-4" />
                    </Button>
                    Configure Service
                  </div>
              ) : isEditing ? "Edit Service" : view === "templates" ? "Select Service Template" : "Configure Service"}
          </DialogTitle>
          <DialogDescription>
            {showDiff ? "Review the changes before applying them." :
             view === "templates" ? "Choose a template to quickly configure a popular service, or start from scratch." : "Configure the upstream service details."}
          </DialogDescription>
        </DialogHeader>

        {showDiff && serviceToEdit && pendingConfig ? (
            <div className="flex flex-col gap-4">
                <ServiceConfigDiff original={serviceToEdit} modified={pendingConfig} />
                <div className="flex justify-end gap-2">
                    <Button variant="outline" onClick={() => setShowDiff(false)}>Back to Edit</Button>
                    <Button onClick={() => performSave(pendingConfig)}>Confirm & Save</Button>
                </div>
            </div>
        ) : view === "templates" ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
                {SERVICE_TEMPLATES.map((template) => {
                    const Icon = template.icon;
                    return (
                        <div
                            key={template.id}
                            className="flex flex-col items-start p-4 border rounded-lg hover:bg-muted/50 cursor-pointer transition-colors"
                            onClick={() => handleTemplateSelect(template)}
                        >
                            <div className="flex items-center gap-3 mb-2">
                                <div className="p-2 bg-primary/10 rounded-md text-primary">
                                    <Icon className="h-5 w-5" />
                                </div>
                                <h3 className="font-semibold">{template.name}</h3>
                            </div>
                            <p className="text-sm text-muted-foreground">{template.description}</p>
                        </div>
                    );
                })}
            </div>
        ) : (
            <Tabs defaultValue="basic" className="w-full">
                <TabsList className="grid w-full grid-cols-3">
                    <TabsTrigger value="basic">Basic Configuration</TabsTrigger>
                    <TabsTrigger value="auth">Authentication</TabsTrigger>
                    <TabsTrigger value="advanced">Advanced (JSON)</TabsTrigger>
                </TabsList>

                <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4 mt-4">

                <TabsContent value="basic" className="space-y-4">
                    <FormField
                    control={form.control}
                    name="name"
                    render={({ field }) => (
                        <FormItem>
                        <FormLabel>Service Name</FormLabel>
                        <FormControl>
                            <Input placeholder="my-service" {...field} />
                        </FormControl>
                        <FormMessage />
                        </FormItem>
                    )}
                    />
                    <FormField
                    control={form.control}
                    name="type"
                    render={({ field }) => (
                        <FormItem>
                        <FormLabel>Service Type</FormLabel>
                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                            <FormControl>
                            <SelectTrigger>
                                <SelectValue placeholder="Select type" />
                            </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                            <SelectItem value="http">HTTP</SelectItem>
                            <SelectItem value="grpc">gRPC</SelectItem>
                            <SelectItem value="command_line">Command Line</SelectItem>
                            <SelectItem value="openapi">OpenAPI</SelectItem>
                            <SelectItem value="other">Other / Advanced</SelectItem>
                            </SelectContent>
                        </Select>
                        <FormMessage />
                        </FormItem>
                    )}
                    />

                    {(selectedType === 'http' || selectedType === 'grpc' || selectedType === 'openapi') && (
                        <FormField
                        control={form.control}
                        name="address"
                        render={({ field }) => (
                            <FormItem>
                            <FormLabel>Address / URL</FormLabel>
                            <FormControl>
                                <Input placeholder="https://api.example.com" {...field} />
                            </FormControl>
                            <FormMessage />
                            </FormItem>
                        )}
                        />
                    )}

                    <FormField
                    control={form.control}
                    name="tags"
                    render={({ field }) => (
                        <FormItem>
                        <FormLabel>Tags</FormLabel>
                        <FormControl>
                            <Input placeholder="prod, external, db (comma separated)" {...field} />
                        </FormControl>
                        <FormDescription>
                            Tags to organize and filter services.
                        </FormDescription>
                        <FormMessage />
                        </FormItem>
                    )}
                    />

                    {selectedType === 'command_line' && (
                        <FormField
                        control={form.control}
                        name="command"
                        render={({ field }) => (
                            <FormItem>
                            <FormLabel>Command</FormLabel>
                            <FormControl>
                                <Input placeholder="python script.py" {...field} />
                            </FormControl>
                            <FormDescription>
                                The command to run the MCP server. Ensure all dependencies are installed.
                            </FormDescription>
                            <FormMessage />
                            </FormItem>
                        )}
                        />
                    )}
                    {selectedType === 'other' && (
                        <div className="text-sm text-muted-foreground">
                            Please switch to the Advanced tab to configure other service types using JSON.
                        </div>
                    )}

                    {/* Hint for Env Vars if using a template that needs them */}
                    {selectedTemplate && selectedTemplate.config.commandLineService?.env && Object.keys(selectedTemplate.config.commandLineService.env).length > 0 && (
                         <div className="p-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-900 rounded-md text-sm text-yellow-800 dark:text-yellow-200">
                            <strong>Note:</strong> This template uses environment variables (e.g., keys).
                            You may need to configure them in the "Advanced (JSON)" tab or ensure they are set in the server environment.
                        </div>
                    )}
                </TabsContent>

                <TabsContent value="auth" className="space-y-4">
                    <div className="space-y-4 border p-4 rounded-md bg-muted/50">
                        <h3 className="text-sm font-medium">Load from Credential</h3>
                        <p className="text-sm text-muted-foreground">Select a saved credential to apply its authentication configuration to this service.</p>
                        <div className="flex gap-2 items-center">
                            <Select onValueChange={handleCredentialSelect}>
                                <SelectTrigger className="w-[300px]">
                                    <SelectValue placeholder="Select credential..." />
                                </SelectTrigger>
                                <SelectContent>
                                    {credentials.map((cred) => (
                                        <SelectItem key={cred.id} value={cred.id}>{cred.name}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                            <Button type="button" variant="ghost" size="icon" onClick={loadCredentials} title="Refresh Credentials">
                                <RotateCw className="h-4 w-4" />
                            </Button>
                        </div>
                    </div>

                    <div className="space-y-2">
                        <h3 className="text-sm font-medium">Current Configuration</h3>
                        {form.watch("upstreamAuth") ? (
                            <div className="text-sm border p-2 rounded">
                                <pre className="whitespace-pre-wrap break-all">
                                    {JSON.stringify(form.watch("upstreamAuth"), null, 2)}
                                </pre>
                                <Button type="button" variant="outline" size="sm" className="mt-2" onClick={() => form.setValue("upstreamAuth", undefined)}>Clear Authentication</Button>
                            </div>
                        ) : (
                            <div className="text-sm text-muted-foreground italic">No authentication configured.</div>
                        )}

                        {/* Interactive OAuth Button */}
                        {((form.watch("upstreamAuth")?.oauth2) || (selectedTemplate?.config?.upstreamAuth?.oauth2)) && (
                            <div className="pt-2">
                                <Button
                                    type="button"
                                    variant="secondary"
                                    className="w-full"
                                    onClick={async () => {
                                        const oauthConfig = form.watch("upstreamAuth")?.oauth2 || selectedTemplate?.config?.upstreamAuth?.oauth2;
                                        if (!oauthConfig) {
                                            toast({ title: "No OAuth Config", description: "Please configure OAuth2 first.", variant: "destructive" });
                                            return;
                                        }

                                        // Save context to session storage so callback page knows what to do
                                        const state = Math.random().toString(36).substring(7); // Temporary client-side state for session key?
                                        // Wait, backend generates state.
                                        // We need to call initiate first.

                                        try {
                                            // If we are editing an existing service, use serviceId.
                                            // If new service, we can't easily associate without ID.
                                            // Limitation: Must save service first?
                                            // Or use Credential flow?
                                            // Let's assume editing for now.

                                            if (!serviceToEdit?.id && !form.getValues("upstreamAuth")) {
                                                 toast({ title: "Save Service First", description: "Please register the service before authenticating.", variant: "default" });
                                                 return;
                                            }

                                            // If we have a selected Credential, we can auth that Credential?
                                            // But the UI logic for selecting credential just copies the config.
                                            // It doesn't link it by ID in the UpstreamServiceConfig (it copies the AuthConfig).
                                            // Unless we change that behavior.

                                            // Let's rely on service_id if editing.
                                            // If creating new, warn user.
                                            const serviceId = serviceToEdit?.id;
                                            if (!serviceId) {
                                                 toast({ title: "Save Required", description: "Please save the service before authenticating.", variant: "destructive" });
                                                 return;
                                            }

                                            const redirectUrl = `${window.location.origin}/oauth/callback`;
                                            const res = await apiClient.initiateOAuth(serviceId, redirectUrl);

                                            // Store context for callback
                                            sessionStorage.setItem(`oauth_pending_${res.state}`, JSON.stringify({
                                                serviceId: serviceId,
                                                redirectUrl: redirectUrl
                                            }));

                                            // Redirect
                                            window.location.href = res.authorization_url;

                                        } catch (e: any) {
                                            console.error(e);
                                            toast({ title: "Failed to Initiate OAuth", description: e.message || "Unknown error", variant: "destructive" });
                                        }
                                    }}
                                >
                                    Authenticate with Provider
                                </Button>
                                <p className="text-xs text-muted-foreground mt-1 text-center">
                                    You will be redirected to the provider to login.
                                </p>
                            </div>
                        )}
                    </div>
                </TabsContent>

                <TabsContent value="advanced">
                    <FormField
                    control={form.control}
                    name="configJson"
                    render={({ field }) => (
                        <FormItem>
                        <FormLabel>Configuration JSON</FormLabel>
                        <FormControl>
                            <Textarea className="font-mono" rows={15} {...field} />
                        </FormControl>
                        <FormDescription>
                            Full JSON configuration for the UpstreamServiceConfig.
                        </FormDescription>
                        <FormMessage />
                        </FormItem>
                    )}
                    />
                </TabsContent>

                <DialogFooter>
                <Button type="submit">{isEditing ? "Review Changes" : "Register Service"}</Button>
                </DialogFooter>
                </form>
                </Form>
            </Tabs>
        )}
      </DialogContent>
    </Dialog>
  );
}
