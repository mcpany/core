/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from "./client";

/**
 * Defines the mode for handling secrets in configurations.
 *
 * @remarks
 * - `'redact'`: Replaces secrets with `'<REDACTED>'`.
 * - `'template'`: Replaces secrets with template placeholders (e.g., `'${API_KEY}'`).
 * - `'unsafe'`: Leaves secrets as plain text (use with caution).
 */
export type SecretHandlingMode = 'redact' | 'template' | 'unsafe';

/**
 * Sanitizes a service configuration by redacting or templating potential secrets.
 *
 * This function creates a deep copy of the configuration and modifies sensitive fields
 * based on the provided mode.
 *
 * @param service - The service configuration to sanitize.
 * @param mode - The secret handling mode (`'redact'`, `'template'`, or `'unsafe'`).
 * @returns A sanitized copy of the configuration.
 */
export function sanitizeServiceConfig(service: UpstreamServiceConfig, mode: SecretHandlingMode): UpstreamServiceConfig {
    // Deep clone to avoid mutating original
    const clone = JSON.parse(JSON.stringify(service));

    if (mode === 'unsafe') {
        return clone;
    }

    // Process commandLineService.env
    if (clone.commandLineService && clone.commandLineService.env) {
        const env = clone.commandLineService.env;
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
        if (clone.upstreamAuth.apiKey) {
             if (mode === 'redact') clone.upstreamAuth.apiKey = '<REDACTED>';
             else if (mode === 'template') clone.upstreamAuth.apiKey = '${API_KEY}';
        }
        if (clone.upstreamAuth.token) {
             if (mode === 'redact') clone.upstreamAuth.token = '<REDACTED>';
             else if (mode === 'template') clone.upstreamAuth.token = '${TOKEN}';
        }
        if (clone.upstreamAuth.basicAuth) {
             if (clone.upstreamAuth.basicAuth.password) {
                 if (mode === 'redact') clone.upstreamAuth.basicAuth.password = '<REDACTED>';
                 else if (mode === 'template') clone.upstreamAuth.basicAuth.password = '${PASSWORD}';
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
