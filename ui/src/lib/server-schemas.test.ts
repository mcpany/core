/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from 'vitest';
import { enrichCommunityServerConfig } from './server-schemas';
import { CommunityServer } from './marketplace-service';

describe('Server Schemas', () => {
    it('should enrich Cloudflare server', () => {
        const server: CommunityServer = {
            category: 'Web',
            name: 'Cloudflare',
            url: 'https://github.com/cloudflare/mcp-server-cloudflare',
            description: 'Cloudflare MCP Server',
            tags: []
        };

        const config = enrichCommunityServerConfig(server);
        expect(config.configurationSchema).toContain('CLOUDFLARE_API_TOKEN');
        expect(config.configurationSchema).toContain('CLOUDFLARE_ACCOUNT_ID');
    });

    it('should enrich Postgres server', () => {
        const server: CommunityServer = {
            category: 'Database',
            name: 'PostgreSQL',
            url: 'https://github.com/modelcontextprotocol/server-postgres',
            description: 'Postgres MCP Server',
            tags: []
        };

        const config = enrichCommunityServerConfig(server);
        expect(config.configurationSchema).toContain('POSTGRES_URL');
    });

    it('should enrich GitHub server', () => {
        const server: CommunityServer = {
            category: 'Dev Tools',
            name: 'GitHub',
            url: 'https://github.com/modelcontextprotocol/server-github',
            description: 'GitHub MCP Server',
            tags: []
        };

        const config = enrichCommunityServerConfig(server);
        expect(config.configurationSchema).toContain('GITHUB_PERSONAL_ACCESS_TOKEN');
    });

    it('should default to empty schema for unknown server', () => {
        const server: CommunityServer = {
            category: 'Other',
            name: 'Unknown Server',
            url: 'https://github.com/example/unknown',
            description: 'Unknown MCP Server',
            tags: []
        };

        const config = enrichCommunityServerConfig(server);
        expect(config.configurationSchema).toBe('');
    });
});
