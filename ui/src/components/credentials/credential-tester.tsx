/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Authentication } from "@proto/config/v1/auth";
import { apiClient, TestAuthRequest, TestAuthResponse } from "@/lib/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Loader2, Play, Globe, AlertTriangle, CheckCircle2, XCircle } from "lucide-react";
import { JsonView } from "@/components/ui/json-view";
import { useToast } from "@/hooks/use-toast";

interface CredentialTesterProps {
    /**
     * The authentication configuration to test.
     * This should come from the form state.
     */
    authConfig: Authentication;
    /**
     * The ID of the credential if it exists (for using stored tokens).
     */
    credentialId?: string;
    /**
     * Pre-filled target URL.
     */
    defaultUrl?: string;
}

/**
 * A component to test authentication credentials against a target URL.
 * Provides detailed diagnostics including status, headers, and body.
 */
export function CredentialTester({ authConfig, credentialId, defaultUrl = "" }: CredentialTesterProps) {
    const [url, setUrl] = useState(defaultUrl);
    const [method, setMethod] = useState("GET");
    const [loading, setLoading] = useState(false);
    const [response, setResponse] = useState<TestAuthResponse | null>(null);
    const { toast } = useToast();

    const handleRunTest = async () => {
        if (!url) {
            toast({
                variant: "destructive",
                title: "Validation Error",
                description: "Please enter a target URL to test.",
            });
            return;
        }

        if (!url.startsWith("http")) {
             toast({
                variant: "destructive",
                title: "Validation Error",
                description: "URL must start with http:// or https://",
            });
            return;
        }

        setLoading(true);
        setResponse(null);

        try {
            const req: TestAuthRequest = {
                target_url: url,
                method: method,
                // If we have a credential ID, send it to use stored secrets/tokens.
                // Otherwise send the auth config directly (e.g. for new credentials).
                // Note: If using authConfig directly, secrets like "${SECRET}" might not be resolved by the test endpoint
                // unless the endpoint handles expansion or we expand them client-side (bad practice).
                // The backend testAuthHandler handles expansion if we pass the Auth object?
                // The backend generally expects resolved values or handles expansion if using the standard authenticator.
                // However, `prepareAndExecuteRequest` creates an authenticator from the config.
                // If the config contains `${...}`, the authenticator might fail or send literal string depending on implementation.
                // Ideally, we should use `credential_id` if available.
                // If creating a NEW credential, we only have `authConfig`.
                // In that case, we hope the user entered the real value or the backend expands it.
                // `AuthManager` usually expands secrets when loading from storage.
                // Here we are passing raw config.
                // Let's pass both if possible, but the API takes one or the other usually?
                // Looking at `server/pkg/app/api_credential.go`:
                // `if req.CredentialID != "" { ... } else { ... }`
                // So it prioritizes ID.

                // Fix: Prioritize authConfig (form values) if present, to ensure we test what the user sees/edits.
                // If we send credential_id, the backend ignores authentication payload.
                credential_id: authConfig ? undefined : credentialId,
                authentication: authConfig
            };

            const res = await apiClient.testAuth(req);
            setResponse(res);

            if (res.error) {
                 toast({
                    variant: "destructive",
                    title: "Test Failed",
                    description: res.error,
                });
            } else if (res.status >= 200 && res.status < 300) {
                toast({
                    title: "Test Successful",
                    description: `Received ${res.status} ${res.status_text}`,
                });
            } else {
                 toast({
                    variant: "default", // Warning-like
                    title: "Test Completed",
                    description: `Received ${res.status} ${res.status_text}`,
                });
            }

        } catch (e: any) {
            console.error("Test failed", e);
            toast({
                variant: "destructive",
                title: "Error",
                description: e.message || "Failed to execute test.",
            });
        } finally {
            setLoading(false);
        }
    };

    const getStatusColor = (status: number) => {
        if (status >= 200 && status < 300) return "text-green-500 border-green-500/50 bg-green-500/10";
        if (status >= 400 && status < 500) return "text-amber-500 border-amber-500/50 bg-amber-500/10";
        if (status >= 500) return "text-red-500 border-red-500/50 bg-red-500/10";
        return "text-muted-foreground";
    };

    return (
        <div className="space-y-4 rounded-lg border p-4 bg-muted/10">
            <div className="flex items-center justify-between">
                <h4 className="text-sm font-medium flex items-center gap-2">
                    <Globe className="h-4 w-4" /> Connection Tester
                </h4>
            </div>

            <div className="flex gap-2">
                <Select value={method} onValueChange={setMethod}>
                    <SelectTrigger className="w-[100px]">
                        <SelectValue placeholder="Method" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="GET">GET</SelectItem>
                        <SelectItem value="POST">POST</SelectItem>
                        <SelectItem value="HEAD">HEAD</SelectItem>
                    </SelectContent>
                </Select>
                <div className="flex-1 relative">
                    <Input
                        placeholder="https://api.example.com/v1/user"
                        value={url}
                        onChange={(e) => setUrl(e.target.value)}
                        className="w-full font-mono text-sm"
                    />
                </div>
                <Button onClick={handleRunTest} disabled={loading} className="w-[120px]">
                    {loading ? (
                        <><Loader2 className="mr-2 h-4 w-4 animate-spin" /> Testing</>
                    ) : (
                        <><Play className="mr-2 h-4 w-4" /> Run Test</>
                    )}
                </Button>
            </div>

            {response && (
                <div className="animate-in fade-in slide-in-from-top-2 duration-300">
                    <div className="flex items-center gap-4 mb-2">
                        <Badge variant="outline" className={`font-mono text-xs px-2 py-1 ${getStatusColor(response.status)}`}>
                            {response.status} {response.status_text}
                        </Badge>
                        {response.error && (
                            <span className="text-xs text-red-500 flex items-center gap-1">
                                <AlertTriangle className="h-3 w-3" /> {response.error}
                            </span>
                        )}
                        {!response.error && response.status === 200 && (
                             <span className="text-xs text-green-500 flex items-center gap-1">
                                <CheckCircle2 className="h-3 w-3" /> Connection Verified
                            </span>
                        )}
                    </div>

                    <Tabs defaultValue="body" className="w-full">
                        <TabsList className="h-8">
                            <TabsTrigger value="body" className="text-xs h-7">Response Body</TabsTrigger>
                            <TabsTrigger value="headers" className="text-xs h-7">Headers</TabsTrigger>
                        </TabsList>
                        <TabsContent value="body" className="mt-2">
                            <Card className="bg-[#1e1e1e] border-none shadow-inner">
                                <CardContent className="p-0">
                                    <JsonView
                                        data={response.body}
                                        maxHeight={300}
                                        className="bg-transparent"
                                    />
                                </CardContent>
                            </Card>
                        </TabsContent>
                        <TabsContent value="headers" className="mt-2">
                            <Card className="bg-muted/30 shadow-inner">
                                <CardContent className="p-2">
                                    <div className="text-xs font-mono grid grid-cols-[1fr_2fr] gap-x-4 gap-y-1">
                                        {response.headers && Object.entries(response.headers).map(([key, value]) => (
                                            <div key={key} className="contents border-b border-muted/50 last:border-0">
                                                <span className="font-semibold text-muted-foreground truncate" title={key}>{key}:</span>
                                                <span className="break-all text-foreground/80">{value}</span>
                                            </div>
                                        ))}
                                        {(!response.headers || Object.keys(response.headers).length === 0) && (
                                            <span className="col-span-2 text-muted-foreground italic">No headers received.</span>
                                        )}
                                    </div>
                                </CardContent>
                            </Card>
                        </TabsContent>
                    </Tabs>
                </div>
            )}
        </div>
    );
}
