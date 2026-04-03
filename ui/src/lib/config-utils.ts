/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from "./client";

/**
 * SecretHandlingMode represents the public SecretHandlingMode entity.
 *
 * Summary: Defines the structured data model representing a handling mode.
 *
 * Parameters:
 *   - None.
 *
 * Returns:
 *   - None.
 *
 * Throws/Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
 */
export type SecretHandlingMode = 'redact' | 'template' | 'unsafe';

/**
 * sanitizeServiceConfig serves as a public interface for interacting with sanitizeServiceConfig.
 *
 * Summary: Sanitize the service config appropriately based on current system conditions.
 *
 * Parameters:
 *   - Refer to the function signature for strongly-typed input arguments.
 *
 * Returns:
 *   - Returns the expected domain model or execution state.
 *
 * Throws/Errors:
 *   - Propagates exceptions from underlying validation layers.
 *
 * Side Effects:
 *   - May mutate state or perform network I/O depending on implementation.
 */
export function sanitizeServiceConfig(service: UpstreamServiceConfig, mode: SecretHandlingMode): UpstreamServiceConfig {
    // Deep clone to avoid mutating original
    const clone = structuredClone(service);

    if (mode === 'unsafe') {
        return clone;
    }

    // Process commandLineService.env
    if (clone.commandLineService && clone.commandLineService.env) {
        const env = clone.commandLineService.env as Record<string, any>;
        for (const key in env) {
            if (Object.prototype.hasOwnProperty.call(env, key)) {
                if (isSecretKey(key)) {
                    if (mode === 'redact') {
                        env[key] = '<REDACTED>';
                    } else if (mode === 'template') {
                        env[key] = `\${${key}}`;
                    }
                }
            }
        }
    }

    // Process upstreamAuth (if it exists in the type, though client.ts shows mapping logic, let's be safe)
    // Based on client.ts: upstreamAuth: s.upstream_auth
    // Looking at proto/config/v1/auth.proto (implied), it might have apiKey, basicAuth etc.
    // If upstreamAuth exists and has fields like 'apiKey', 'token', we should redact them.
    if (clone.upstreamAuth) {
        const auth = clone.upstreamAuth as any;
        if (auth.apiKey) {
             if (mode === 'redact') auth.apiKey = '<REDACTED>';
             else if (mode === 'template') auth.apiKey = '${API_KEY}';
        }
        if (auth.token) {
             if (mode === 'redact') auth.token = '<REDACTED>';
             else if (mode === 'template') auth.token = '${TOKEN}';
        }
        if (auth.basicAuth) {
             if (auth.basicAuth.password) {
                 if (mode === 'redact') auth.basicAuth.password = '<REDACTED>';
                 else if (mode === 'template') auth.basicAuth.password = '${PASSWORD}';
             }
        }
    }

    return clone;
}

function isSecretKey(key: string): boolean {
    const upper = key.toUpperCase();
    return (
        upper.includes('KEY') ||
        upper.includes('SECRET') ||
        upper.includes('TOKEN') ||
        upper.includes('PASSWORD') ||
        upper.includes('PWD') ||
        upper.includes('AUTH') ||
        upper.includes('CREDENTIAL')
    );
}
