/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Loader2, Play, CheckCircle2, XCircle, AlertTriangle, Terminal, Globe, ShieldCheck } from "lucide-react";
import { cn } from "@/lib/utils";
import { ScrollArea } from "@/components/ui/scroll-area";

interface CredentialTesterProps {
    credentialId?: string;
    credentialName?: string;
}

/**
 * Component for testing credentials against services.
 * @param props - The component props.
 * @param props.credentialId - The ID of the credential to test.
 * @param props.credentialName - The name of the credential.
 * @returns The rendered component.
 */
export function CredentialTester({ credentialId, credentialName }: CredentialTesterProps) {
    const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
    const [selectedService, setSelectedService] = useState<string>("custom");
    const [customUrl, setCustomUrl] = useState("");
    const [loading, setLoading] = useState(false);
    const [result, setResult] = useState<{ success: boolean; message: string } | null>(null);
    const [loadingServices, setLoadingServices] = useState(false);

    useEffect(() => {
        const loadServices = async () => {
            setLoadingServices(true);
            try {
                const list = await apiClient.listServices();
                // Filter for services that might use credentials (HTTP/MCP)
                const compatible = list.filter(s => s.httpService || s.mcpService);
                setServices(compatible);
            } catch (e) {
                console.error("Failed to load services for tester", e);
            } finally {
                setLoadingServices(false);
            }
        };
        loadServices();
    }, []);

    const handleTest = async () => {
        if (!credentialId) return;

        setLoading(true);
        setResult(null);

        try {
            // Determine config to send
            let serviceType = "HTTP";
            let serviceConfig: any = {};

            if (selectedService === "custom") {
                if (!customUrl) {
                    setLoading(false);
                    return; // Should validate
                }
                // Construct a temporary HTTP service config for the test endpoint
                serviceType = "HTTP";
                serviceConfig = {
                    httpService: {
                        address: customUrl
                    }
                };
            } else {
                const svc = services.find(s => s.name === selectedService);
                if (svc) {
                    if (svc.httpService) {
                        serviceType = "HTTP";
                        serviceConfig = { httpService: svc.httpService };
                    } else if (svc.mcpService) {
                        // Assuming MCP uses HTTP transport for auth test if available?
                        // Or just pass the whole config and let backend decide
                        // Backend supports HTTP and CMD.
                        // Let's pass the whole object structure expected by protojson
                        serviceConfig = svc;
                        // But backend `Unmarshal` expects `UpstreamServiceConfig`.
                        // We need to map our client-side camelCase to snake_case or rely on protojson forgivingness?
                        // `protojson` expects snake_case for field names usually if not using `UseProtoNames`.
                        // Actually our client object is camelCase. We might need to transform it.
                        // However, for "HTTP" type, I constructed a simple object above.
                        // For real service, let's try to extract the URL and use "HTTP" type manually if possible to be safe.

                        if (svc.httpService) {
                             serviceConfig = { http_service: { address: svc.httpService.address } };
                        } else if (svc.mcpService?.httpConnection) {
                             serviceConfig = { mcp_service: { http_connection: { http_address: svc.mcpService.httpConnection.httpAddress } } };
                        }
                    }
                }
            }

            const res = await apiClient.testAuth({
                credential_id: credentialId,
                service_type: serviceType,
                service_config: serviceConfig
            });

            setResult(res);
        } catch (e: any) {
            setResult({
                success: false,
                message: e.message || "Test failed"
            });
        } finally {
            setLoading(false);
        }
    };

    if (!credentialId) {
        return (
            <div className="rounded-md border border-dashed p-8 text-center bg-muted/20">
                <ShieldCheck className="mx-auto h-8 w-8 text-muted-foreground opacity-50 mb-2" />
                <h3 className="font-medium text-sm text-muted-foreground">Save Credential to Test</h3>
                <p className="text-xs text-muted-foreground mt-1">
                    For security reasons, credentials must be encrypted and stored before they can be verified against upstream services.
                </p>
            </div>
        );
    }

    return (
        <Card className="border-muted shadow-sm">
            <CardHeader className="pb-3 bg-muted/10 border-b">
                <CardTitle className="text-sm font-medium flex items-center gap-2">
                    <ActivityIcon className="h-4 w-4 text-primary" />
                    Connection Diagnostics
                </CardTitle>
            </CardHeader>
            <CardContent className="p-4 space-y-4">
                <div className="grid gap-4">
                    <div className="space-y-2">
                        <Label className="text-xs font-medium text-muted-foreground uppercase">Target Service</Label>
                        <Select value={selectedService} onValueChange={setSelectedService}>
                            <SelectTrigger>
                                <SelectValue placeholder="Select a service" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="custom">
                                    <span className="flex items-center gap-2">
                                        <Globe className="h-3 w-3 text-muted-foreground" />
                                        Custom URL
                                    </span>
                                </SelectItem>
                                {loadingServices ? (
                                    <div className="p-2 text-xs text-muted-foreground flex items-center justify-center">
                                        <Loader2 className="h-3 w-3 animate-spin mr-2" /> Loading services...
                                    </div>
                                ) : (
                                    services.map(s => (
                                        <SelectItem key={s.name} value={s.name}>
                                            <span className="flex items-center gap-2">
                                                <div className="w-2 h-2 rounded-full bg-blue-500/50" />
                                                {s.name}
                                            </span>
                                        </SelectItem>
                                    ))
                                )}
                            </SelectContent>
                        </Select>
                    </div>

                    {selectedService === "custom" && (
                        <div className="space-y-2 animate-in fade-in slide-in-from-top-2">
                            <Label className="text-xs font-medium text-muted-foreground uppercase">Test URL</Label>
                            <div className="relative">
                                <Globe className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                                <Input
                                    value={customUrl}
                                    onChange={(e) => setCustomUrl(e.target.value)}
                                    placeholder="https://api.example.com/v1/user"
                                    className="pl-9 font-mono text-sm"
                                />
                            </div>
                            <p className="text-[10px] text-muted-foreground">
                                Enter a safe, read-only endpoint (e.g. <code>/user</code> or <code>/health</code>) to verify authentication.
                            </p>
                        </div>
                    )}

                    <Button
                        onClick={handleTest}
                        disabled={loading || (selectedService === "custom" && !customUrl)}
                        className="w-full"
                    >
                        {loading ? (
                            <>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                Testing Connection...
                            </>
                        ) : (
                            <>
                                <Play className="mr-2 h-4 w-4" />
                                Run Verification
                            </>
                        )}
                    </Button>
                </div>

                {result && (
                    <div className={cn(
                        "rounded-md border p-4 animate-in fade-in slide-in-from-bottom-2",
                        result.success ? "bg-green-500/10 border-green-500/20" : "bg-red-500/10 border-red-500/20"
                    )}>
                        <div className="flex items-start gap-3">
                            <div className={cn(
                                "shrink-0 p-1 rounded-full",
                                result.success ? "bg-green-500 text-white" : "bg-red-500 text-white"
                            )}>
                                {result.success ? <CheckCircle2 className="h-4 w-4" /> : <XCircle className="h-4 w-4" />}
                            </div>
                            <div className="flex-1 space-y-1">
                                <p className={cn("text-sm font-semibold", result.success ? "text-green-700 dark:text-green-300" : "text-red-700 dark:text-red-300")}>
                                    {result.success ? "Authentication Successful" : "Authentication Failed"}
                                </p>
                                <div className="text-xs font-mono bg-background/50 p-2 rounded border border-black/5 dark:border-white/5 whitespace-pre-wrap break-all">
                                    {result.message}
                                </div>
                            </div>
                        </div>
                    </div>
                )}
            </CardContent>
        </Card>
    );
}

function ActivityIcon(props: any) {
    return (
      <svg
        {...props}
        xmlns="http://www.w3.org/2000/svg"
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
      </svg>
    )
  }
