/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { apiClient, Credential } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Loader2, Play, CheckCircle2, AlertTriangle, ShieldCheck, Globe } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";

interface CredentialTesterProps {
    credential: Partial<Credential>;
}

export function CredentialTester({ credential }: CredentialTesterProps) {
    const [targetUrl, setTargetUrl] = useState("");
    const [method, setMethod] = useState("GET");
    const [isRunning, setIsRunning] = useState(false);
    const [result, setResult] = useState<{ success: boolean; message: string } | null>(null);

    const handleTest = async () => {
        if (!targetUrl) return;

        setIsRunning(true);
        setResult(null);

        try {
            // Construct a transient UpstreamServiceConfig
            // The backend expects a Service Config that has http_service or similar.
            const serviceConfig = {
                httpService: {
                    address: targetUrl
                }
            };

            const res = await apiClient.testAuth({
                credential: credential as Credential,
                serviceType: "HTTP",
                serviceConfig: serviceConfig
            });

            setResult({
                success: res.success,
                message: res.message
            });
        } catch (e: any) {
            setResult({
                success: false,
                message: e.message || "Test failed"
            });
        } finally {
            setIsRunning(false);
        }
    };

    // Determine auth type for display
    const authType = credential.authentication?.apiKey ? "API Key" :
                     credential.authentication?.bearerToken ? "Bearer Token" :
                     credential.authentication?.basicAuth ? "Basic Auth" :
                     credential.authentication?.oauth2 ? "OAuth 2.0" : "None";

    const isOAuth = !!credential.authentication?.oauth2;

    return (
        <Card className="border-dashed bg-muted/20">
            <CardHeader className="pb-3">
                <CardTitle className="text-sm font-medium flex items-center gap-2">
                    <ShieldCheck className="h-4 w-4 text-primary" />
                    Test Configuration
                </CardTitle>
                <CardDescription>
                    Verify this credential against a real endpoint.
                </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
                <div className="grid gap-2">
                    <div className="flex gap-2">
                        <Select value={method} onValueChange={setMethod}>
                            <SelectTrigger className="w-[100px]">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="GET">GET</SelectItem>
                                <SelectItem value="POST">POST</SelectItem>
                                <SelectItem value="HEAD">HEAD</SelectItem>
                            </SelectContent>
                        </Select>
                        <div className="relative flex-1">
                            <Globe className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                            <Input
                                placeholder="https://api.example.com/user"
                                value={targetUrl}
                                onChange={(e) => setTargetUrl(e.target.value)}
                                className="pl-9"
                            />
                        </div>
                        <Button onClick={handleTest} disabled={isRunning || !targetUrl || isOAuth}>
                            {isRunning ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
                            <span className="ml-2 hidden sm:inline">Test</span>
                        </Button>
                    </div>
                    <p className="text-[10px] text-muted-foreground">
                        We will send a {method} request to this URL using the configured <strong>{authType}</strong>.
                        {isOAuth && <span className="block text-amber-500 mt-1">OAuth 2.0 testing requires the full connection flow. Please save and connect first.</span>}
                    </p>
                </div>

                {result && (
                    <div className={`rounded-md border p-4 text-sm ${result.success ? "bg-green-500/10 border-green-500/20" : "bg-red-500/10 border-red-500/20"}`}>
                        <div className="flex items-start gap-3">
                            {result.success ? (
                                <CheckCircle2 className="h-5 w-5 text-green-500 shrink-0 mt-0.5" />
                            ) : (
                                <AlertTriangle className="h-5 w-5 text-red-500 shrink-0 mt-0.5" />
                            )}
                            <div className="space-y-1 overflow-hidden">
                                <p className="font-medium">{result.success ? "Connection Successful" : "Connection Failed"}</p>
                                <p className="text-muted-foreground break-all whitespace-pre-wrap">{result.message}</p>
                            </div>
                        </div>
                    </div>
                )}
            </CardContent>
        </Card>
    );
}
