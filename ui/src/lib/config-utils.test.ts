/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { sanitizeServiceConfig } from './config-utils';
import { UpstreamServiceConfig } from './client';

describe('sanitizeServiceConfig', () => {
    const baseMockService: UpstreamServiceConfig = {
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
    } as UpstreamServiceConfig;

    it('should redact secrets when mode is redact', () => {
        const sanitizedRedact = sanitizeServiceConfig(baseMockService, 'redact');
        expect(sanitizedRedact.commandLineService?.env?.['API_KEY']).toBe('<REDACTED>');
        expect(sanitizedRedact.commandLineService?.env?.['DB_PASSWORD']).toBe('<REDACTED>');
        expect(sanitizedRedact.commandLineService?.env?.['AUTH_TOKEN']).toBe('<REDACTED>');
        expect(sanitizedRedact.commandLineService?.env?.['PUBLIC_VAR']).toBe('public-value');
    });

    it('should template secrets when mode is template', () => {
        const sanitizedTemplate = sanitizeServiceConfig(baseMockService, 'template');
        expect(sanitizedTemplate.commandLineService?.env?.['API_KEY']).toBe('${API_KEY}');
        expect(sanitizedTemplate.commandLineService?.env?.['DB_PASSWORD']).toBe('${DB_PASSWORD}');
        expect(sanitizedTemplate.commandLineService?.env?.['AUTH_TOKEN']).toBe('${AUTH_TOKEN}');
        expect(sanitizedTemplate.commandLineService?.env?.['PUBLIC_VAR']).toBe('public-value');
    });

    it('should keep secrets when mode is unsafe', () => {
        const sanitizedUnsafe = sanitizeServiceConfig(baseMockService, 'unsafe');
        expect(sanitizedUnsafe.commandLineService?.env?.['API_KEY']).toBe('super-secret-key');
        expect(sanitizedUnsafe.commandLineService?.env?.['DB_PASSWORD']).toBe('super-secret-password');
        expect(sanitizedUnsafe.commandLineService?.env?.['AUTH_TOKEN']).toBe('jwt-token');
        expect(sanitizedUnsafe.commandLineService?.env?.['PUBLIC_VAR']).toBe('public-value');
    });

    it('should handle missing env', () => {
        const serviceNoEnv = { ...baseMockService, commandLineService: { command: 'cmd' } };
        const sanitizedMissing = sanitizeServiceConfig(serviceNoEnv as UpstreamServiceConfig, 'redact');
        expect(sanitizedMissing.commandLineService?.env).toBeUndefined();
    });
});
