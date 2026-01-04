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
import { Plus } from "lucide-react";

const formSchema = z.object({
  name: z.string().min(1, "Name is required"),
  type: z.enum(["grpc", "http", "command_line", "openapi", "other"]),
  address: z.string().optional(),
  command: z.string().optional(),
  configJson: z.string().optional(), // For advanced mode
});

interface RegisterServiceDialogProps {
  onSuccess?: () => void;
  trigger?: React.ReactNode;
  serviceToEdit?: UpstreamServiceConfig;
}

export function RegisterServiceDialog({ onSuccess, trigger, serviceToEdit }: RegisterServiceDialogProps) {
  const [open, setOpen] = useState(false);
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
  } : {
      name: "",
      type: "http" as const,
      address: "",
      command: "",
      configJson: "{\n  \"name\": \"my-service\",\n  \"httpService\": {\n    \"address\": \"https://api.example.com\"\n  }\n}",
  };

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues,
  });

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
              id: serviceToEdit?.id,
          } as any;

          if (values.type === 'grpc') {
              (config as any).grpcService = { address: values.address || "" };
          } else if (values.type === 'http') {
              (config as any).httpService = { address: values.address || "" };
          } else if (values.type === 'command_line') {
              (config as any).commandLineService = { command: values.command || "" };
          } else if (values.type === 'openapi') {
               (config as any).openapiService = { address: values.address || "" };
          }
      }

      if (isEditing) {
          await apiClient.updateService(config as any);
           toast({ title: "Service Updated", description: `${values.name} updated successfully.` });
      } else {
          await apiClient.registerService(config as any);
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
            <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="basic">Basic Configuration</TabsTrigger>
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
