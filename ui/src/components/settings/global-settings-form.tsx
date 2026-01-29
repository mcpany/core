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
  // Rate Limit
  rate_limit_enabled: z.boolean(),
  rate_limit_rps: z.coerce.number().min(0.1),
  rate_limit_burst: z.coerce.number().min(1),
  // Telemetry
  telemetry_traces: z.enum(["otlp", "stdout", "none"]),
  telemetry_metrics: z.enum(["otlp", "stdout", "none"]),
  telemetry_endpoint: z.string().optional(),
  // Alerts
  alerts_enabled: z.boolean(),
  alerts_webhook: z.string().optional(),
});

type SettingsValues = z.infer<typeof settingsSchema>;

/**
 * GlobalSettingsForm component.
 * @returns The rendered component.
 */
export function GlobalSettingsForm() {
  const [loading, setLoading] = useState(false);
  const [isReadOnly, setIsReadOnly] = useState(false);

  const form = useForm<SettingsValues>({
    resolver: zodResolver(settingsSchema),
    defaultValues: {
      mcp_listen_address: ":8080",
      log_level: "INFO",
      log_format: "text",
      audit_enabled: true,
      dlp_enabled: false,
      gc_interval: "1h",
      rate_limit_enabled: false,
      rate_limit_rps: 10,
      rate_limit_burst: 20,
      telemetry_traces: "none",
      telemetry_metrics: "none",
      telemetry_endpoint: "",
      alerts_enabled: false,
      alerts_webhook: "",
    },
  });

  useEffect(() => {
    async function loadSettings() {
      try {
        const settings = await apiClient.getGlobalSettings();
        if (settings) {
            const mapLogLevel = (l: any) => {
                if (l === 4 || l === "DEBUG" || l === "LOG_LEVEL_DEBUG") return "DEBUG";
                if (l === 3 || l === "ERROR" || l === "LOG_LEVEL_ERROR") return "ERROR";
                if (l === 2 || l === "WARN" || l === "LOG_LEVEL_WARN") return "WARN";
                return "INFO";
            };
            const mapLogFormat = (f: any) => {
                if (f === 2 || f === "json" || f === "LOG_FORMAT_JSON") return "json";
                return "text";
            };

            // Map backend enum integers to strings if necessary
            form.reset({
                mcp_listen_address: settings.mcp_listen_address || ":8080",
                log_level: mapLogLevel(settings.log_level),
                log_format: mapLogFormat(settings.log_format),
                audit_enabled: settings.audit?.enabled || false,
                dlp_enabled: settings.dlp?.enabled || false,
                gc_interval: settings.gc_settings?.interval || "1h",
                rate_limit_enabled: settings.rate_limit?.is_enabled || false,
                rate_limit_rps: settings.rate_limit?.requests_per_second || 10,
                rate_limit_burst: settings.rate_limit?.burst || 20,
                telemetry_traces: settings.telemetry?.traces_exporter || "none",
                telemetry_metrics: settings.telemetry?.metrics_exporter || "none",
                telemetry_endpoint: settings.telemetry?.otlp_endpoint || "",
                alerts_enabled: settings.alerts?.enabled || false,
                alerts_webhook: settings.alerts?.webhook_url || "",
            });
            if (settings.read_only) {
                setIsReadOnly(true);
            }
        }
      } catch (e) {
        console.error("Failed to load settings", e);
      }
    }
    loadSettings();
  }, [form]);

  async function onSubmit(data: SettingsValues) {
    if (isReadOnly) return;
    setLoading(true);
    try {
       // Map strings back to backend enums
       const payload = {
           mcp_listen_address: data.mcp_listen_address,
           log_level: data.log_level === "DEBUG" ? 4 : data.log_level === "ERROR" ? 3 : data.log_level === "WARN" ? 2 : 1,
           log_format: data.log_format === "json" ? 2 : 1,
           audit: { enabled: data.audit_enabled },
           dlp: { enabled: data.dlp_enabled },
           gc_settings: { interval: data.gc_interval },
           rate_limit: {
               is_enabled: data.rate_limit_enabled,
               requests_per_second: data.rate_limit_rps,
               burst: data.rate_limit_burst
           },
           telemetry: {
               traces_exporter: data.telemetry_traces,
               metrics_exporter: data.telemetry_metrics,
               otlp_endpoint: data.telemetry_endpoint
           },
           alerts: {
               enabled: data.alerts_enabled,
               webhook_url: data.alerts_webhook
           }
       };
       await apiClient.saveGlobalSettings(payload);
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
      <CardContent className="space-y-6">
        {isReadOnly && (
            <div className="bg-yellow-50 dark:bg-yellow-900/20 border-l-4 border-yellow-500 p-4 text-yellow-700 dark:text-yellow-400">
                <p className="font-bold">Configuration Read-Only</p>
                <p className="text-sm">Additional configuration can be found in the file configuration. This configuration cannot be edited via the UI.</p>
            </div>
        )}
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            <fieldset disabled={isReadOnly} className="space-y-6 group-disabled:opacity-50">
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
                        <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value} disabled={isReadOnly}>
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
                        <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value} disabled={isReadOnly}>
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
                        disabled={isReadOnly}
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
                        disabled={isReadOnly}
                        />
                    </FormControl>
                    </FormItem>
                )}
                />

                {/* Rate Limiting Section */}
                <div className="space-y-4 pt-4 border-t">
                    <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Rate Limiting</h3>
                    <FormField
                        control={form.control}
                        name="rate_limit_enabled"
                        render={({ field }) => (
                            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                            <div className="space-y-0.5">
                                <FormLabel className="text-base">Global Rate Limit</FormLabel>
                                <FormDescription>
                                Limit the number of requests per second globally.
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
                    {form.watch("rate_limit_enabled") && (
                        <div className="grid grid-cols-2 gap-4 pl-4 border-l-2">
                            <FormField
                                control={form.control}
                                name="rate_limit_rps"
                                render={({ field }) => (
                                    <FormItem>
                                    <FormLabel>Requests Per Second</FormLabel>
                                    <FormControl>
                                        <Input type="number" step="0.1" {...field} />
                                    </FormControl>
                                    <FormMessage />
                                    </FormItem>
                                )}
                            />
                            <FormField
                                control={form.control}
                                name="rate_limit_burst"
                                render={({ field }) => (
                                    <FormItem>
                                    <FormLabel>Burst</FormLabel>
                                    <FormControl>
                                        <Input type="number" step="1" {...field} />
                                    </FormControl>
                                    <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </div>
                    )}
                </div>

                {/* Telemetry Section */}
                <div className="space-y-4 pt-4 border-t">
                    <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Telemetry & Monitoring</h3>
                    <div className="grid grid-cols-2 gap-4">
                        <FormField
                            control={form.control}
                            name="telemetry_traces"
                            render={({ field }) => (
                                <FormItem>
                                <FormLabel>Traces Exporter</FormLabel>
                                <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value} disabled={isReadOnly}>
                                    <FormControl>
                                    <SelectTrigger>
                                        <SelectValue placeholder="Select exporter" />
                                    </SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                    <SelectItem value="none">None</SelectItem>
                                    <SelectItem value="otlp">OTLP (OpenTelemetry)</SelectItem>
                                    <SelectItem value="stdout">Stdout (Console)</SelectItem>
                                    </SelectContent>
                                </Select>
                                <FormMessage />
                                </FormItem>
                            )}
                        />
                        <FormField
                            control={form.control}
                            name="telemetry_metrics"
                            render={({ field }) => (
                                <FormItem>
                                <FormLabel>Metrics Exporter</FormLabel>
                                <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value} disabled={isReadOnly}>
                                    <FormControl>
                                    <SelectTrigger>
                                        <SelectValue placeholder="Select exporter" />
                                    </SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                    <SelectItem value="none">None</SelectItem>
                                    <SelectItem value="otlp">OTLP (OpenTelemetry)</SelectItem>
                                    <SelectItem value="stdout">Stdout (Console)</SelectItem>
                                    </SelectContent>
                                </Select>
                                <FormMessage />
                                </FormItem>
                            )}
                        />
                    </div>
                    {(form.watch("telemetry_traces") === "otlp" || form.watch("telemetry_metrics") === "otlp") && (
                        <FormField
                            control={form.control}
                            name="telemetry_endpoint"
                            render={({ field }) => (
                                <FormItem>
                                <FormLabel>OTLP Endpoint</FormLabel>
                                <FormControl>
                                    <Input placeholder="localhost:4317" {...field} />
                                </FormControl>
                                <FormDescription>gRPC endpoint for OpenTelemetry Collector</FormDescription>
                                <FormMessage />
                                </FormItem>
                            )}
                        />
                    )}
                </div>

                {/* Alerts Section */}
                <div className="space-y-4 pt-4 border-t">
                    <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Alerting</h3>
                    <FormField
                        control={form.control}
                        name="alerts_enabled"
                        render={({ field }) => (
                            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                            <div className="space-y-0.5">
                                <FormLabel className="text-base">System Alerts</FormLabel>
                                <FormDescription>
                                Receive notifications for critical errors and health issues.
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
                    {form.watch("alerts_enabled") && (
                        <FormField
                            control={form.control}
                            name="alerts_webhook"
                            render={({ field }) => (
                                <FormItem>
                                <FormLabel>Webhook URL</FormLabel>
                                <FormControl>
                                    <Input placeholder="https://hooks.slack.com/..." {...field} />
                                </FormControl>
                                <FormMessage />
                                </FormItem>
                            )}
                        />
                    )}
                </div>
            </div>
            </fieldset>

            <Button type="submit" disabled={loading || isReadOnly}>
                {loading ? "Saving..." : isReadOnly ? "Read Only" : "Save Settings"}
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
