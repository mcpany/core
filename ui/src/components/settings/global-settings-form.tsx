/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
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

const settingsSchema = z.object({
  mcp_listen_address: z.string().min(1, "Address is required"),
  log_level: z.enum(["INFO", "WARN", "ERROR", "DEBUG"]),
  log_format: z.enum(["text", "json"]),
  audit_enabled: z.boolean(),
  dlp_enabled: z.boolean(),
  gc_interval: z.string(),
});

type SettingsValues = z.infer<typeof settingsSchema>;

export function GlobalSettingsForm() {
  const [loading, setLoading] = useState(false);

  const form = useForm<SettingsValues>({
    resolver: zodResolver(settingsSchema),
    defaultValues: {
      mcp_listen_address: ":8080",
      log_level: "INFO",
      log_format: "text",
      audit_enabled: true,
      dlp_enabled: false,
      gc_interval: "1h",
    },
  });

  useEffect(() => {
    async function loadSettings() {
      try {
        const settings = await apiClient.getGlobalSettings();
        if (settings) {
            // Map backend enum integers to strings if necessary
            form.reset({
                mcp_listen_address: settings.mcpListenAddress || ":8080",
                log_level: settings.logLevel === 4 ? "DEBUG" : settings.logLevel === 3 ? "ERROR" : settings.logLevel === 2 ? "WARN" : "INFO",
                log_format: settings.logFormat === 2 ? "json" : "text",
                audit_enabled: settings.audit?.enabled || false,
                dlp_enabled: settings.dlp?.enabled || false,
                gc_interval: settings.gcSettings?.interval || "1h",
            });
        }
      } catch (e) {
        console.error("Failed to load settings", e);
      }
    }
    loadSettings();
  }, [form]);

  async function onSubmit(data: SettingsValues) {
    setLoading(true);
    try {
       // Map strings back to backend enums
       const payload = {
           mcpListenAddress: data.mcp_listen_address,
           logLevel: data.log_level === "DEBUG" ? 4 : data.log_level === "ERROR" ? 3 : data.log_level === "WARN" ? 2 : 1,
           logFormat: data.log_format === "json" ? 2 : 1,
           audit: { enabled: data.audit_enabled },
           dlp: { enabled: data.dlp_enabled },
           gcSettings: { interval: data.gc_interval }
       };
       await apiClient.saveGlobalSettings(payload as any);
    } catch (e) {
      console.error("Failed to save settings", e);
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card className="backdrop-blur-sm bg-background/50">
      <CardHeader>
        <CardTitle>Global Configuration</CardTitle>
        <CardDescription>Server-wide operational parameters.</CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <FormField
                control={form.control}
                name="mcp_listen_address"
                render={({ field }) => (
                    <FormItem>
                    <FormLabel>Listen Address</FormLabel>
                    <FormControl>
                        <Input placeholder=":8080" {...field} />
                    </FormControl>
                    <FormDescription>
                        The address the MCP server listens on.
                    </FormDescription>
                    <FormMessage />
                    </FormItem>
                )}
                />

                <FormField
                control={form.control}
                name="gc_interval"
                render={({ field }) => (
                    <FormItem>
                    <FormLabel>GC Interval</FormLabel>
                    <FormControl>
                        <Input placeholder="1h" {...field} />
                    </FormControl>
                    <FormDescription>
                        Frequency of garbage collection.
                    </FormDescription>
                    <FormMessage />
                    </FormItem>
                )}
                />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                 <FormField
                    control={form.control}
                    name="log_level"
                    render={({ field }) => (
                        <FormItem>
                        <FormLabel>Log Level</FormLabel>
                        <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value}>
                            <FormControl>
                            <SelectTrigger>
                                <SelectValue placeholder="Select a level" />
                            </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                            <SelectItem value="DEBUG">DEBUG</SelectItem>
                            <SelectItem value="INFO">INFO</SelectItem>
                            <SelectItem value="WARN">WARN</SelectItem>
                            <SelectItem value="ERROR">ERROR</SelectItem>
                            </SelectContent>
                        </Select>
                        <FormMessage />
                        </FormItem>
                    )}
                    />
                     <FormField
                    control={form.control}
                    name="log_format"
                    render={({ field }) => (
                        <FormItem>
                        <FormLabel>Log Format</FormLabel>
                        <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value}>
                            <FormControl>
                            <SelectTrigger>
                                <SelectValue placeholder="Select a format" />
                            </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                            <SelectItem value="text">Text</SelectItem>
                            <SelectItem value="json">JSON</SelectItem>
                            </SelectContent>
                        </Select>
                        <FormMessage />
                        </FormItem>
                    )}
                    />
            </div>

            <div className="flex flex-col space-y-4">
                 <FormField
                control={form.control}
                name="audit_enabled"
                render={({ field }) => (
                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                    <div className="space-y-0.5">
                        <FormLabel className="text-base">Audit Logging</FormLabel>
                        <FormDescription>
                        Enable detailed audit logs for all operations.
                        </FormDescription>
                    </div>
                    <FormControl>
                        <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                        />
                    </FormControl>
                    </FormItem>
                )}
                />

                 <FormField
                control={form.control}
                name="dlp_enabled"
                render={({ field }) => (
                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                    <div className="space-y-0.5">
                        <FormLabel className="text-base">Data Loss Prevention (DLP)</FormLabel>
                        <FormDescription>
                        Redact sensitive information from logs and outputs.
                        </FormDescription>
                    </div>
                    <FormControl>
                        <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                        />
                    </FormControl>
                    </FormItem>
                )}
                />
            </div>

            <Button type="submit" disabled={loading}>
                {loading ? "Saving..." : "Save Settings"}
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
