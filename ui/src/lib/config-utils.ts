/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from "./client";

/**
 * Intent: Document SecretHandlingMode
 *
 * Params:
 *   - None
 *
 * Returns:
 *   - None
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * Defines the mode for handling secrets in configurations.
 * - 'redact': Replaces secrets with '<REDACTED>'.
 * - 'template': Replaces secrets with template placeholders (e.g., '${API_KEY}').
 * - 'unsafe': Leaves secrets as plain text (use with caution).
 */
export type SecretHandlingMode = 'redact' | 'template' | 'unsafe';

/**
 * Intent: Document sanitizeServiceConfig
 *
 * Params:
 *   - None
 *
 * Sanitizes a service configuration by redacting or templating potential secrets.
 *
 * Summary: Redacts or templates secrets embedded in a service configuration object.
 *
 * Parameters:
 *   - service (UpstreamServiceConfig): The service configuration to sanitize.
 *   - mode (SecretHandlingMode): The secret handling mode ('redact', 'template', or 'unsafe').
 *
 * Returns:
 *   - UpstreamServiceConfig: A sanitized deep copy of the configuration.
 *
 * Errors:
 *   - None.
 *
 * Side Effects:
 *   - None.
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
