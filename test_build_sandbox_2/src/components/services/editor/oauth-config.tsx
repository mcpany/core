/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Authentication } from "@/lib/client";

interface OAuthConfigProps {
    auth: Authentication["oauth2"];
    onChange: (auth: Authentication["oauth2"]) => void;
}

/**
 * OAuthConfig component.
 * @param props - The component props.
 * @param props.auth - The authentication configuration.
 * @param props.onChange - Callback function when value changes.
 * @returns The rendered component.
 */
export function OAuthConfig({ auth, onChange }: OAuthConfigProps) {
    const updateAuth = (updates: Partial<NonNullable<Authentication["oauth2"]>>) => {
        onChange({ ...auth, ...updates } as any);
    };

    // Helper to safely get plain text value from SecretValue
    const getSecretValue = (secret: any): string => {
        if (!secret) return "";
        // Check for generated proto structure { value: { plainText: "..." } }
        if (secret.value && typeof secret.value === 'object' && 'plainText' in secret.value) {
            return secret.value.plainText;
        }
        // Check for simplified structure { plainText: "..." } (sometimes seen in simplified TS clients)
        if (typeof secret.plainText === 'string') {
            return secret.plainText;
        }
        // Fallback: empty string or log warning
        return "";
    };

    const clientId = getSecretValue(auth?.clientId);
    const clientSecret = getSecretValue(auth?.clientSecret);

    return (
        <div className="space-y-4 border-l-2 border-primary/20 pl-4">
             <div className="space-y-2">
                <Label htmlFor="client-id">Client ID</Label>
                <Input
                    id="client-id"
                    value={clientId}
                    onChange={(e) => updateAuth({ clientId: { value: { plainText: e.target.value } } as any })}
                    placeholder="Enter Client ID"
                />
            </div>
            <div className="space-y-2">
                <Label htmlFor="client-secret">Client Secret</Label>
                <Input
                    id="client-secret"
                    type="password"
                    value={clientSecret}
                    onChange={(e) => updateAuth({ clientSecret: { value: { plainText: e.target.value } } as any })}
                    placeholder="Enter Client Secret"
                />
            </div>
            <div className="space-y-2">
                <Label htmlFor="auth-url">Authorization URL</Label>
                <Input
                    id="auth-url"
                    value={auth?.authorizationUrl || ""}
                    onChange={(e) => updateAuth({ authorizationUrl: e.target.value })}
                    placeholder="https://github.com/login/oauth/authorize"
                />
            </div>
            <div className="space-y-2">
                <Label htmlFor="token-url">Token URL</Label>
                <Input
                    id="token-url"
                    value={auth?.tokenUrl || ""}
                    onChange={(e) => updateAuth({ tokenUrl: e.target.value })}
                    placeholder="https://github.com/login/oauth/access_token"
                />
            </div>
            <div className="space-y-2">
                <Label htmlFor="scopes">Scopes</Label>
                <Input
                    id="scopes"
                    value={auth?.scopes || ""}
                    onChange={(e) => updateAuth({ scopes: e.target.value })}
                    placeholder="read write repo"
                />
                <p className="text-xs text-muted-foreground">Space separated list of scopes.</p>
            </div>
        </div>
    );
}
