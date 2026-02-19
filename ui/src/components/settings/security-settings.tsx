/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { useForm, useFieldArray } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { apiClient } from "@/lib/client";
import { Trash, Plus, Shield, Gauge, Activity } from "lucide-react";
import { toast } from "sonner";

const securitySchema = z.object({
  // DLP
  dlp_enabled: z.boolean(),
  dlp_patterns: z.array(z.object({
    value: z.string().min(1, "Pattern cannot be empty")
  })),

  // Rate Limiting
  ratelimit_enabled: z.boolean(),
  ratelimit_rps: z.coerce.number().min(0),
  ratelimit_burst: z.coerce.number().min(0),

  // Audit
  audit_enabled: z.boolean(),
  audit_storage: z.enum(["file", "sqlite", "postgres", "webhook", "splunk", "datadog"]),
  audit_path: z.string().optional(),
  audit_log_args: z.boolean(),
  audit_log_results: z.boolean(),
});

type SecurityValues = z.infer<typeof securitySchema>;
type StorageType = "file" | "sqlite" | "postgres" | "webhook" | "splunk" | "datadog";

export function SecuritySettings() {
  const [loading, setLoading] = useState(false);
  const [isReadOnly, setIsReadOnly] = useState(false);

  const form = useForm<SecurityValues>({
    resolver: zodResolver(securitySchema),
    defaultValues: {
      dlp_enabled: false,
      dlp_patterns: [],
      ratelimit_enabled: false,
      ratelimit_rps: 20,
      ratelimit_burst: 50,
      audit_enabled: true,
      audit_storage: "file",
      audit_path: "audit.log",
      audit_log_args: false,
      audit_log_results: false,
    },
  });

  const { fields: dlpFields, append: appendDlp, remove: removeDlp } = useFieldArray({
    control: form.control,
    name: "dlp_patterns",
  });

  useEffect(() => {
    async function loadSettings() {
      try {
        const settings = await apiClient.getGlobalSettings();
        if (settings) {
            let storageType: StorageType = "file";
            const sType = settings.audit?.storage_type;

            if (sType === 1 || sType === "STORAGE_TYPE_FILE") storageType = "file";
            else if (sType === 2 || sType === "STORAGE_TYPE_SQLITE") storageType = "sqlite";
            else if (sType === 3 || sType === "STORAGE_TYPE_POSTGRES") storageType = "postgres";
            else if (sType === 4 || sType === "STORAGE_TYPE_WEBHOOK") storageType = "webhook";
            else if (sType === 5 || sType === "STORAGE_TYPE_SPLUNK") storageType = "splunk";
            else if (sType === 6 || sType === "STORAGE_TYPE_DATADOG") storageType = "datadog";

            form.reset({
                dlp_enabled: settings.dlp?.enabled || false,
                dlp_patterns: (settings.dlp?.custom_patterns || []).map((p: string) => ({ value: p })),

                ratelimit_enabled: settings.rate_limit?.is_enabled || false,
                ratelimit_rps: settings.rate_limit?.requests_per_second || 20,
                ratelimit_burst: settings.rate_limit?.burst || 50,

                audit_enabled: settings.audit?.enabled || false,
                audit_storage: storageType,
                audit_path: settings.audit?.output_path || "",
                audit_log_args: settings.audit?.log_arguments || false,
                audit_log_results: settings.audit?.log_results || false,
            });
            if (settings.read_only) {
                setIsReadOnly(true);
            }
        }
      } catch (e) {
        console.error("Failed to load settings", e);
        toast.error("Failed to load settings");
      }
    }
    loadSettings();
  }, [form]);

  async function onSubmit(data: SecurityValues) {
    if (isReadOnly) return;
    setLoading(true);
    try {
        const current = await apiClient.getGlobalSettings();

        const storageTypeMap: Record<string, number> = {
            "file": 1,
            "sqlite": 2,
            "postgres": 3,
            "webhook": 4,
            "splunk": 5,
            "datadog": 6
        };

        const payload = {
            ...current,
            dlp: {
                ...current.dlp,
                enabled: data.dlp_enabled,
                custom_patterns: data.dlp_patterns.map(p => p.value)
            },
            rate_limit: {
                ...current.rate_limit,
                is_enabled: data.ratelimit_enabled,
                requests_per_second: data.ratelimit_rps,
                burst: data.ratelimit_burst,
            },
            audit: {
                ...current.audit,
                enabled: data.audit_enabled,
                storage_type: storageTypeMap[data.audit_storage],
                output_path: data.audit_path,
                log_arguments: data.audit_log_args,
                log_results: data.audit_log_results,
            }
        };

        await apiClient.saveGlobalSettings(payload);
        toast.success("Security settings saved");
    } catch (e) {
      console.error("Failed to save settings", e);
      toast.error("Failed to save settings");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">

        {/* DLP Section */}
        <Card className="backdrop-blur-sm bg-background/50">
            <CardHeader>
                <div className="flex items-center gap-2">
                    <Shield className="h-5 w-5 text-primary" />
                    <CardTitle>Data Loss Prevention (DLP)</CardTitle>
                </div>
                <CardDescription>Configure redaction rules for sensitive information in logs and traces.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
                <FormField
                    control={form.control}
                    name="dlp_enabled"
                    render={({ field }) => (
                        <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                            <div className="space-y-0.5">
                                <FormLabel className="text-base">Enable DLP</FormLabel>
                                <FormDescription>
                                    Automatically redact Email, Credit Card, and SSN patterns.
                                </FormDescription>
                            </div>
                            <FormControl>
                                <Switch
                                    checked={field.value}
                                    onCheckedChange={field.onChange}
                                    disabled={isReadOnly}
                                />
                            </FormControl>
                        </FormItem>
                    )}
                />

                {form.watch("dlp_enabled") && (
                    <div className="space-y-2 border rounded-lg p-4 bg-muted/20">
                        <div className="flex justify-between items-center">
                            <FormLabel className="text-sm font-medium">Custom Redaction Patterns (Regex)</FormLabel>
                            <Button type="button" variant="outline" size="sm" onClick={() => appendDlp({ value: "" })} disabled={isReadOnly}>
                                <Plus className="h-4 w-4 mr-1" /> Add Pattern
                            </Button>
                        </div>
                        {dlpFields.length === 0 && <p className="text-sm text-muted-foreground italic">No custom patterns configured.</p>}
                        {dlpFields.map((field, index) => (
                            <div key={field.id} className="flex gap-2">
                                <FormField
                                    control={form.control}
                                    name={`dlp_patterns.${index}.value`}
                                    render={({ field }) => (
                                        <FormItem className="flex-1">
                                            <FormControl>
                                                <Input {...field} placeholder="e.g. sk-[a-zA-Z0-9]+" disabled={isReadOnly} />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                <Button type="button" variant="ghost" size="icon" onClick={() => removeDlp(index)} disabled={isReadOnly}>
                                    <Trash className="h-4 w-4 text-destructive" />
                                </Button>
                            </div>
                        ))}
                    </div>
                )}
            </CardContent>
        </Card>

        {/* Rate Limiting Section */}
        <Card className="backdrop-blur-sm bg-background/50">
            <CardHeader>
                <div className="flex items-center gap-2">
                    <Gauge className="h-5 w-5 text-primary" />
                    <CardTitle>Global Rate Limiting</CardTitle>
                </div>
                <CardDescription>Protect the server from abuse by limiting request volume.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
                <FormField
                    control={form.control}
                    name="ratelimit_enabled"
                    render={({ field }) => (
                        <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                            <div className="space-y-0.5">
                                <FormLabel className="text-base">Enable Rate Limiting</FormLabel>
                                <FormDescription>
                                    Enforce global request limits across all users and services.
                                </FormDescription>
                            </div>
                            <FormControl>
                                <Switch
                                    checked={field.value}
                                    onCheckedChange={field.onChange}
                                    disabled={isReadOnly}
                                />
                            </FormControl>
                        </FormItem>
                    )}
                />

                <div className="grid grid-cols-2 gap-4">
                    <FormField
                        control={form.control}
                        name="ratelimit_rps"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Requests Per Second (RPS)</FormLabel>
                                <FormControl>
                                    <Input type="number" {...field} disabled={!form.watch("ratelimit_enabled") || isReadOnly} />
                                </FormControl>
                                <FormDescription>Target throughput.</FormDescription>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                    <FormField
                        control={form.control}
                        name="ratelimit_burst"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Burst Capacity</FormLabel>
                                <FormControl>
                                    <Input type="number" {...field} disabled={!form.watch("ratelimit_enabled") || isReadOnly} />
                                </FormControl>
                                <FormDescription>Allowed short-term spikes.</FormDescription>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                </div>
            </CardContent>
        </Card>

        {/* Audit Section */}
        <Card className="backdrop-blur-sm bg-background/50">
            <CardHeader>
                <div className="flex items-center gap-2">
                    <Activity className="h-5 w-5 text-primary" />
                    <CardTitle>Audit & Observability</CardTitle>
                </div>
                <CardDescription>Configure detailed request logging for compliance.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
                <FormField
                    control={form.control}
                    name="audit_enabled"
                    render={({ field }) => (
                        <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                            <div className="space-y-0.5">
                                <FormLabel className="text-base">Enable Audit Logging</FormLabel>
                                <FormDescription>
                                    Log every tool execution and configuration change.
                                </FormDescription>
                            </div>
                            <FormControl>
                                <Switch
                                    checked={field.value}
                                    onCheckedChange={field.onChange}
                                    disabled={isReadOnly}
                                />
                            </FormControl>
                        </FormItem>
                    )}
                />

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <FormField
                        control={form.control}
                        name="audit_storage"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Storage Backend</FormLabel>
                                <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value} disabled={!form.watch("audit_enabled") || isReadOnly}>
                                    <FormControl>
                                        <SelectTrigger>
                                            <SelectValue placeholder="Select storage" />
                                        </SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                        <SelectItem value="file">Local File</SelectItem>
                                        <SelectItem value="sqlite">SQLite</SelectItem>
                                        <SelectItem value="postgres">PostgreSQL</SelectItem>
                                        <SelectItem value="webhook">Webhook</SelectItem>
                                        {/* Splunk/Datadog hidden for now as they require complex config */}
                                    </SelectContent>
                                </Select>
                                <FormMessage />
                            </FormItem>
                        )}
                    />

                    {form.watch("audit_storage") === "file" && (
                        <FormField
                            control={form.control}
                            name="audit_path"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Log File Path</FormLabel>
                                    <FormControl>
                                        <Input {...field} placeholder="/var/log/mcpany/audit.log" disabled={!form.watch("audit_enabled") || isReadOnly} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />
                    )}
                </div>

                <div className="flex gap-4">
                    <FormField
                        control={form.control}
                        name="audit_log_args"
                        render={({ field }) => (
                            <FormItem className="flex items-center space-x-2">
                                <FormControl>
                                    <Switch checked={field.value} onCheckedChange={field.onChange} disabled={!form.watch("audit_enabled") || isReadOnly} />
                                </FormControl>
                                <FormLabel className="font-normal">Log Arguments</FormLabel>
                            </FormItem>
                        )}
                    />
                    <FormField
                        control={form.control}
                        name="audit_log_results"
                        render={({ field }) => (
                            <FormItem className="flex items-center space-x-2">
                                <FormControl>
                                    <Switch checked={field.value} onCheckedChange={field.onChange} disabled={!form.watch("audit_enabled") || isReadOnly} />
                                </FormControl>
                                <FormLabel className="font-normal">Log Results</FormLabel>
                            </FormItem>
                        )}
                    />
                </div>
            </CardContent>
        </Card>

        <Button type="submit" disabled={loading || isReadOnly} className="w-full sm:w-auto">
            {loading ? "Saving..." : isReadOnly ? "Read Only" : "Save Security Settings"}
        </Button>
      </form>
    </Form>
  );
}
