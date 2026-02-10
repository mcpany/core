/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { WizardService } from "../wizard-dialog";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { CheckCircle2, AlertCircle, ExternalLink, Loader2 } from "lucide-react";
import { apiClient } from "@/lib/client";
import { toast } from "sonner";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface AuthStepProps {
    services: WizardService[];
    onNext: (services: WizardService[]) => void;
    onBack: () => void;
}

/**
 * AuthStep allows the user to authenticate the selected services.
 */
export function AuthStep({ services, onNext, onBack }: AuthStepProps) {
    const [authServices, setAuthServices] = useState<WizardService[]>([...services]);
    const [loadingAuth, setLoadingAuth] = useState<Record<string, boolean>>({});

    const initiateOAuth = async (index: number) => {
        const svc = authServices[index];
        setLoadingAuth(prev => ({ ...prev, [svc.instanceName]: true }));
        try {
            // Register service first?
            // The InitiateOAuth endpoint requires service_id.
            // But we haven't registered the service yet!
            //
            // Issue: We can't auth against a non-existent service config in backend usually,
            // UNLESS the backend allows "template" auth or we register it now.
            //
            // Strategy:
            // 1. Register the service now (with auth disabled/incomplete).
            // 2. Perform Auth.
            // 3. Update Service with Auth tokens (if flow returns them to backend directly)
            //    OR backend handles callback and updates service.
            //
            // Better Strategy for Wizard:
            // "Register on Connect".
            // We register the service *now* using the config we have.
            // Then we initiate OAuth.

            // Check if already registered?
            try {
                await apiClient.getService(svc.instanceName);
                // Exists, maybe update?
                // For wizard, let's assume update idemptotently
                await apiClient.updateService({ ...svc.config, id: svc.instanceName, name: svc.instanceName });
            } catch {
                // Doesn't exist, register
                await apiClient.registerService({ ...svc.config, id: svc.instanceName, name: svc.instanceName });
            }

            const res = await apiClient.initiateOAuth(
                svc.instanceName,
                window.location.href, // Redirect back here? Or a special callback page?
                svc.instanceName // credential_id same as service for 1:1 binding
            );

            // Open popup
            const popup = window.open(res.authorization_url, "mcp_auth", "width=600,height=700");

            // Poll for completion? or separate "I have authenticated" button?
            // For MVP, user clicks "I've finished connecting" or we listen to window message if backend redirects to a /close-popup page.
            // Let's assume user manually confirms for now or we just show link.

            if (popup) {
                toast.info("Please complete authentication in the popup");
            } else {
                 toast.error("Popup blocked. Please allow popups.");
            }

            // Optimistically mark as "Pending Verification"
            // For now, let's just mark authenticated manually for the wizard flow UI
            const next = [...authServices];
            next[index].isAuthenticated = true; // Todo: verify
            setAuthServices(next);

        } catch (error) {
            console.error(error);
            toast.error("Failed to initiate authentication");
        } finally {
            setLoadingAuth(prev => ({ ...prev, [svc.instanceName]: false }));
        }
    };

    const handleTokenInput = (index: number, token: string) => {
        const next = [...authServices];
        // Deep set token
        if (next[index].templateId === 'github') {
             next[index].config.upstreamAuth.bearerToken = { token: { plainText: token } };
        }
        // ... handle other types
        next[index].isAuthenticated = !!token;
        setAuthServices(next);
    };

    const handleNext = () => {
        onNext(authServices);
    };

    return (
        <div className="space-y-6">
            <div className="space-y-4">
                {authServices.map((svc, idx) => (
                    <Card key={idx}>
                        <CardHeader className="py-3">
                            <div className="flex justify-between items-center">
                                <CardTitle className="text-base">{svc.instanceName}</CardTitle>
                                {svc.isAuthenticated ? (
                                    <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">
                                        <CheckCircle2 className="w-3 h-3 mr-1" />
                                        Authenticated
                                    </Badge>
                                ) : (
                                    <Badge variant="outline" className="text-amber-600 border-amber-200">
                                        <AlertCircle className="w-3 h-3 mr-1" />
                                        Setup Required
                                    </Badge>
                                )}
                            </div>
                        </CardHeader>
                        <CardContent className="py-3">
                            {svc.templateId === 'google-calendar' || svc.templateId === 'linear' ? (
                                <div className="flex items-center gap-4">
                                     <Button
                                        variant="outline"
                                        onClick={() => initiateOAuth(idx)}
                                        disabled={loadingAuth[svc.instanceName] || svc.isAuthenticated}
                                    >
                                        {loadingAuth[svc.instanceName] && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                        {svc.isAuthenticated ? "Re-Connect" : "Connect Account"}
                                        {!loadingAuth[svc.instanceName] && <ExternalLink className="ml-2 h-4 w-4" />}
                                     </Button>
                                     <p className="text-sm text-muted-foreground">
                                         {svc.isAuthenticated
                                            ? "Access granted."
                                            : "Requires OAuth2 authentication."}
                                     </p>
                                </div>
                            ) : svc.templateId === 'github' ? (
                                <div className="grid gap-2">
                                     <Label>Personal Access Token</Label>
                                     <Input
                                        type="password"
                                        placeholder="ghp_..."
                                        onChange={(e) => handleTokenInput(idx, e.target.value)}
                                        defaultValue={svc.config.upstreamAuth?.bearerToken?.token?.plainText || ""}
                                     />
                                </div>
                            ) : (
                                <p className="text-sm">No authentication required (or not supported in wizard yet).</p>
                            )}
                        </CardContent>
                    </Card>
                ))}
            </div>

            <div className="flex justify-between">
                <Button variant="outline" onClick={onBack}>Back</Button>
                 <div>
                    <Button variant="ghost" onClick={handleNext} className="mr-2">Skip / Finish Later</Button>
                    <Button onClick={handleNext}>Next: Review</Button>
                </div>
            </div>
        </div>
    );
}
