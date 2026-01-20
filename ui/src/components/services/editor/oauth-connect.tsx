/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { Loader2, LogIn } from "lucide-react";

interface OAuthConnectProps {
    serviceId: string; // Or name
    serviceName: string;
    isSaved: boolean;
}

/**
 * OAuthConnect component.
 * @param props - The component props.
 * @param props.serviceId - The unique identifier for service.
 * @param props.serviceName - The name of the service.
 * @param props.isSaved - The issaved property.
 * @returns The rendered component.
 */
export function OAuthConnect({ serviceId, serviceName, isSaved }: OAuthConnectProps) {
    const [loading, setLoading] = useState(false);
    const { toast } = useToast();

    const handleConnect = async () => {
        if (!isSaved) {
            toast({
                title: "Save Required",
                description: "Please save the service configuration before connecting.",
                variant: "destructive",
            });
            return;
        }

        setLoading(true);
        try {
            // Determine redirect URL
            const redirectUrl = `${window.location.origin}/oauth/callback`;

            // Initiate flow
            const response = await apiClient.initiateOAuth(serviceId, redirectUrl);

            // Store context for callback
            const state = response.state;
            if (state) {
                sessionStorage.setItem(`oauth_pending_${state}`, JSON.stringify({
                    serviceId: serviceId,
                    serviceName: serviceName,
                    timestamp: Date.now()
                }));
            }

            // Redirect user
            if (response.authorization_url) {
                window.location.href = response.authorization_url;
            } else {
                throw new Error("No authorization URL returned");
            }

        } catch (e: any) {
            console.error("OAuth Init Failed", e);
            toast({
                title: "Connection Failed",
                description: e.message || "Failed to initiate OAuth flow",
                variant: "destructive",
            });
            setLoading(false);
        }
    };

    if (!isSaved) {
        return (
            <div className="p-4 border border-yellow-200 bg-yellow-50 dark:bg-yellow-900/10 rounded text-sm text-yellow-800 dark:text-yellow-200">
                Please save the service to enable OAuth connection.
            </div>
        );
    }

    return (
        <div className="pt-4">
            <Button onClick={handleConnect} disabled={loading} className="w-full sm:w-auto">
                {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <LogIn className="mr-2 h-4 w-4" />}
                Connect with {serviceName}
            </Button>
            <p className="text-xs text-muted-foreground mt-2">
                This will redirect you to the provider to authorize access.
            </p>
        </div>
    );
}
