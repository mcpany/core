/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { sanitizeServiceConfig, SecretHandlingMode } from './config-utils';
import { UpstreamServiceConfig } from './client';

describe('sanitizeServiceConfig', () => {
    const mockService: UpstreamServiceConfig = {
        id: 'test-service',
        name: 'Test Service',
        commandLineService: {
            command: 'test-cmd',
            env: {
                'API_KEY': 'super-secret-key',
                'DB_PASSWORD': 'super-secret-password',
                'PUBLIC_VAR': 'public-value',
                'AUTH_TOKEN': 'jwt-token'
            },
            workingDirectory: '/tmp'
        }
    } as any; // Casting as UpstreamServiceConfig might have other required fields in strict TS

    it('should redact secrets when mode is redact', () => {
        const sanitized = sanitizeServiceConfig(mockService, 'redact');
        expect(sanitized.commandLineService?.env?.['API_KEY']).toBe('<REDACTED>');
        expect(sanitized.commandLineService?.env?.['DB_PASSWORD']).toBe('<REDACTED>');
        expect(sanitized.commandLineService?.env?.['AUTH_TOKEN']).toBe('<REDACTED>');
        expect(sanitized.commandLineService?.env?.['PUBLIC_VAR']).toBe('public-value');
    });

    it('should template secrets when mode is template', () => {
        const sanitized = sanitizeServiceConfig(mockService, 'template');
        expect(sanitized.commandLineService?.env?.['API_KEY']).toBe('${API_KEY}');
        expect(sanitized.commandLineService?.env?.['DB_PASSWORD']).toBe('${DB_PASSWORD}');
        expect(sanitized.commandLineService?.env?.['AUTH_TOKEN']).toBe('${AUTH_TOKEN}');
        expect(sanitized.commandLineService?.env?.['PUBLIC_VAR']).toBe('public-value');
    });

    it('should keep secrets when mode is unsafe', () => {
        const sanitized = sanitizeServiceConfig(mockService, 'unsafe');
        expect(sanitized.commandLineService?.env?.['API_KEY']).toBe('super-secret-key');
        expect(sanitized.commandLineService?.env?.['DB_PASSWORD']).toBe('super-secret-password');
        expect(sanitized.commandLineService?.env?.['AUTH_TOKEN']).toBe('jwt-token');
        expect(sanitized.commandLineService?.env?.['PUBLIC_VAR']).toBe('public-value');
    });

    it('should handle missing env', () => {
        const serviceNoEnv = { ...mockService, commandLineService: { command: 'cmd' } };
        const sanitized = sanitizeServiceConfig(serviceNoEnv as any, 'redact');
        expect(sanitized.commandLineService?.env).toBeUndefined();
    });
});
