/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { AlertCircle, CheckCircle2, ShieldAlert } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

interface OAuthStatusBannerProps {
    status: "connected" | "disconnected" | "expired" | "error";
    serviceName: string;
    onConnect: () => void;
    errorMessage?: string;
}

/**
 * OAuthStatusBanner component.
 * @param props - The component props.
 * @param props.status - The current status.
 * @param props.serviceName - The name of the service.
 * @param props.onConnect - The onConnect property.
 * @param props.errorMessage - The error message or object.
 * @returns The rendered component.
 */
export const OAuthStatusBanner: React.FC<OAuthStatusBannerProps> = ({
    status,
    serviceName,
    onConnect,
    errorMessage,
}) => {
    if (status === "connected") {
        return (
            <Alert className="bg-green-50 border-green-200 dark:bg-green-900/10 dark:border-green-900/20">
                <CheckCircle2 className="h-4 w-4 text-green-600 dark:text-green-400" />
                <AlertTitle className="text-green-800 dark:text-green-300">Connected</AlertTitle>
                <AlertDescription className="text-green-700 dark:text-green-400 text-sm">
                    {serviceName} is authenticated and ready to use.
                </AlertDescription>
            </Alert>
        );
    }

    if (status === "disconnected") {
        return (
            <Alert className="bg-blue-50 border-blue-200 dark:bg-blue-900/10 dark:border-blue-900/20">
                <AlertCircle className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                <AlertTitle className="text-blue-800 dark:text-blue-300">Authentication Required</AlertTitle>
                <AlertDescription className="flex items-center justify-between gap-4">
                    <span className="text-blue-700 dark:text-blue-400 text-sm">
                        Connect your {serviceName} account to enable all features.
                    </span>
                    <Button size="sm" onClick={onConnect} className="h-8">
                        Connect Account
                    </Button>
                </AlertDescription>
            </Alert>
        );
    }

    if (status === "expired") {
        return (
            <Alert variant="destructive" className="bg-yellow-50 border-yellow-200 text-yellow-900 dark:bg-yellow-900/10 dark:border-yellow-900/20 dark:text-yellow-400">
                <ShieldAlert className="h-4 w-4" />
                <AlertTitle>Session Expired</AlertTitle>
                <AlertDescription className="flex items-center justify-between gap-4">
                    <span className="text-sm">Your {serviceName} session has expired. Please reconnect.</span>
                    <Button variant="outline" size="sm" onClick={onConnect} className="h-8 border-yellow-300 hover:bg-yellow-100 dark:border-yellow-900">
                        Reconnect
                    </Button>
                </AlertDescription>
            </Alert>
        );
    }

    return (
        <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Authentication Error</AlertTitle>
            <AlertDescription className="flex items-center justify-between gap-4">
                <span className="text-sm">{errorMessage || `Failed to authenticate with ${serviceName}.`}</span>
                <Button variant="outline" size="sm" onClick={onConnect} className="h-8">
                    Retry
                </Button>
            </AlertDescription>
        </Alert>
    );
};
