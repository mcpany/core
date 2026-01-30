import { describe, it, expect } from 'vitest';
import { resolveServerConfig } from '@/lib/server-registry';
import { UpstreamServiceConfig } from '@/lib/client';

describe('Server Registry', () => {
    it('should resolve known servers by repo URL', () => {
        expect(resolveServerConfig('https://github.com/modelcontextprotocol/server-github')).toBeDefined();
        expect(resolveServerConfig('https://github.com/modelcontextprotocol/server-filesystem')).toBeDefined();
        expect(resolveServerConfig('https://github.com/modelcontextprotocol/server-slack')).toBeDefined();
    });

    it('should resolve known servers by name', () => {
        expect(resolveServerConfig('server-github')).toBeDefined();
        expect(resolveServerConfig('mcp-server-cloudflare')).toBeDefined();
    });

    it('should configure filesystem args correctly', () => {
        const item = resolveServerConfig('server-filesystem');
        expect(item).toBeDefined();

        const baseConfig: UpstreamServiceConfig = {
            id: 'test',
            name: 'test',
            version: '1.0.0',
            commandLineService: {
                command: 'npx -y @modelcontextprotocol/server-filesystem',
                env: {}
            }
        } as any;

        const values = { "ALLOWED_PATHS": "/tmp/test" };
        const result = item!.configure(baseConfig, values);

        expect(result.commandLineService?.command).toContain('/tmp/test');
        expect(result.commandLineService?.command).toBe('npx -y @modelcontextprotocol/server-filesystem /tmp/test');
    });

    it('should configure github env vars correctly', () => {
        const item = resolveServerConfig('server-github');
        expect(item).toBeDefined();

        const baseConfig: UpstreamServiceConfig = {
            id: 'test',
            name: 'test',
            version: '1.0.0',
            commandLineService: {
                command: 'npx -y @modelcontextprotocol/server-github',
                env: {}
            }
        } as any;

        const values = { "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_123" };
        const result = item!.configure(baseConfig, values);

        expect(result.commandLineService?.env?.["GITHUB_PERSONAL_ACCESS_TOKEN"]).toBeDefined();
        expect((result.commandLineService?.env?.["GITHUB_PERSONAL_ACCESS_TOKEN"] as any).plainText).toBe('ghp_123');
    });
});
