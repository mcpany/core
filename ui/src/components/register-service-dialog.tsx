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
import { Plus, RotateCw } from "lucide-react";

const formSchema = z.object({
  name: z.string().min(1, "Name is required"),
  type: z.enum(["grpc", "http", "command_line", "openapi", "other"]),
  address: z.string().optional(),
  command: z.string().optional(),
  configJson: z.string().optional(), // For advanced mode
  upstreamAuth: z.any().optional(), // Store auth config object
});

interface RegisterServiceDialogProps {
  onSuccess?: () => void;
  trigger?: React.ReactNode;
  serviceToEdit?: UpstreamServiceConfig;
}

export function RegisterServiceDialog({ onSuccess, trigger, serviceToEdit }: RegisterServiceDialogProps) {
  const [open, setOpen] = useState(false);
  const [credentials, setCredentials] = useState<Credential[]>([]);
  const { toast } = useToast();
  const isEditing = !!serviceToEdit;

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
  } : {
      name: "",
      type: "http" as const,
      address: "",
      command: "",
      configJson: "{\n  \"name\": \"my-service\",\n  \"httpService\": {\n    \"address\": \"https://api.example.com\"\n  }\n}",
      upstreamAuth: undefined,
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
  // Simple check: load once
  if (open && credentials.length === 0) {
      loadCredentials();
  }

  const onSubmit = async (values: z.infer<typeof formSchema>) => {
    try {
      let config: UpstreamServiceConfig;

      // Logic: If on Advanced tab (we can't easily tell), OR type is 'other', use JSON.
      // But we can check if JSON field was modified significantly?
      // Simpler: If configJson parses to valid object and has matching name, prefer it?
      // Let's assume if type is 'other', use JSON.
      // If type is standard, construct config.

      // However, we want to allow users to use Basic tab for quick setup.
      // Let's assume: If user provided JSON, they probably want to use it.
      // But defaultValues has JSON.

      // Improved logic:
      // If values.type === 'other', strictly use JSON.
      // Else, construct config from basic fields.

      if (values.type === 'other') {
          if (!values.configJson) throw new Error("Config JSON is required for 'Other' type");
          config = JSON.parse(values.configJson);
      } else {
          // Construct config
          config = {
              name: values.name,
              id: serviceToEdit?.id || "",
              version: "1.0.0",
              disable: false,
              priority: 0,
              loadBalancingStrategy: 0,
              sanitizedName: "",
              callPolicies: [],
              preCallHooks: [],
              postCallHooks: [],
              prompts: [],
              autoDiscoverTool: false,
          };

          if (values.type === 'grpc') {
              config.grpcService = { address: values.address || "", useReflection: true, tools: [], resources: [], calls: {}, prompts: [], protoCollection: [], protoDefinitions: [] };
          } else if (values.type === 'http') {
              config.httpService = { address: values.address || "", tools: [], calls: {}, resources: [], prompts: [] };
          } else if (values.type === 'command_line') {
              config.commandLineService = { command: values.command || "", workingDirectory: "", local: false, env: {}, tools: [], resources: [], prompts: [], communicationProtocol: 0, calls: {} };
          } else if (values.type === 'openapi') {
               config.openapiService = { address: values.address || "", tools: [], calls: {}, resources: [], prompts: [] };
          }

          if (values.upstreamAuth) {
              config.upstreamAuth = values.upstreamAuth;
          }
      }

      if (isEditing) {
          await apiClient.updateService(config);
           toast({ title: "Service Updated", description: `${values.name} updated successfully.` });
      } else {
          await apiClient.registerService(config);
          toast({ title: "Service Registered", description: `${values.name} registered successfully.` });
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

  const selectedType = form.watch("type");

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {trigger || <Button><Plus className="mr-2 h-4 w-4" /> Register Service</Button>}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px] max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{isEditing ? "Edit Service" : "Register New Service"}</DialogTitle>
          <DialogDescription>
            Configure an upstream service to be exposed by MCPAny.
          </DialogDescription>
        </DialogHeader>
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
              <Button type="submit">{isEditing ? "Save Changes" : "Register Service"}</Button>
            </DialogFooter>
            </form>
            </Form>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}
