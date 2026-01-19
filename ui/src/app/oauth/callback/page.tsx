/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

function OAuthCallbackContent() {
    const searchParams = useSearchParams();
    const router = useRouter();
    const { toast } = useToast();
    const [status, setStatus] = useState<"processing" | "success" | "error">("processing");
    const [errorMsg, setErrorMsg] = useState("");

    useEffect(() => {
        const code = searchParams.get("code");
        const state = searchParams.get("state");
        const error = searchParams.get("error");

        if (error) {
            setStatus("error");
            setErrorMsg(error);
            return;
        }

        if (!code || !state) {
            setStatus("error");
            setErrorMsg("Missing code or state in callback URL");
            return;
        }

        // Retrieve stored context from sessionStorage
        // We expect the 'state' to match what we stored, or we just trust the flow if we don't strictly validate state in frontend (backend provided it).
        // Use the value stored in sessionStorage to know the serviceID/credentialID.
        // Key: `oauth_pending_${state}`

        const storedContextStr = sessionStorage.getItem(`oauth_pending_${state}`);
        if (!storedContextStr) {
             // Maybe state mismatch?
             // Or maybe user opened in new tab?
             // Proceed with caution or error?
             // Let's try to proceed if we can, but we need serviceID/credentialID.
             // If we don't have them, we can't call backend.
             setStatus("error");
             setErrorMsg("Invalid state or session expired. Please try again.");
             return;
        }

        const context = JSON.parse(storedContextStr);
        // Clean up
        sessionStorage.removeItem(`oauth_pending_${state}`);

        const handleCallback = async () => {
            try {
                // Call backend to exchange code
                // We use the root endpoint or API endpoint? `apiClient` usually uses /api/v1.
                // But `handleOAuthCallback` is available at /auth/oauth/callback too.
                // Let's add a method to `apiClient` to handle this.
                // Or call fetch directly.
                // The API expects POST.

                await apiClient.handleOAuthCallback(
                    context.serviceId,
                    code,
                    window.location.origin + window.location.pathname, // Must match what we sent in initiate
                    context.credentialId
                );

                setStatus("success");
                toast({ title: "Authentication Successful", description: "You can now close this window." });
                // Optional: Close window if opened as popup
                if (window.opener) {
                    setTimeout(() => window.close(), 1500);
                } else {
                    // Redirect to tools or services?
                    setTimeout(() => router.push("/services"), 1500);
                }

            } catch (e: any) {
                console.error("OAuth Exchange Failed", e);
                setStatus("error");
                setErrorMsg(e.message || "Failed to exchange token");
            }
        };

        handleCallback();

    }, [searchParams, router, toast]);

    return (
        <div className="flex flex-col items-center justify-center min-h-screen bg-background">
            <div className="w-full max-w-md p-8 space-y-4 text-center border rounded-lg shadow-lg bg-card text-card-foreground">
                <h1 className="text-2xl font-bold">Authentication</h1>

                {status === "processing" && (
                     <div className="flex flex-col items-center gap-4">
                        <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin" />
                        <p>Completing authentication...</p>
                    </div>
                )}

                {status === "success" && (
                    <div className="space-y-2 text-green-600 dark:text-green-400">
                         <h2 className="text-xl font-semibold">Success!</h2>
                         <p>Your account has been connected.</p>
                         <p className="text-sm text-muted-foreground">Redirecting...</p>
                    </div>
                )}

                {status === "error" && (
                    <div className="space-y-4 text-red-600 dark:text-red-400">
                        <h2 className="text-xl font-semibold">Authentication Failed</h2>
                        <p className="p-2 border border-red-200 bg-red-50 dark:bg-red-900/10 rounded">
                            {errorMsg || "Unknown error occurred"}
                        </p>
                        <button
                            onClick={() => router.push("/services")}
                            className="px-4 py-2 text-sm font-medium text-white bg-primary rounded hover:bg-primary/90"
                        >
                            Back to Services
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
}

export default function OAuthCallbackPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <OAuthCallbackContent />
    </Suspense>
  );
}
