/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from 'vitest';
import { enrichCommunityServerConfig } from './server-schemas';
import { CommunityServer } from '@/lib/marketplace-service';

describe('enrichCommunityServerConfig', () => {
    it('should inject GitHub schema for Official GitHub server', () => {
        const server: CommunityServer = {
            name: 'GitHub',
            url: 'https://github.com/modelcontextprotocol/server-github',
            description: 'GitHub Integration',
            category: 'Dev',
            tags: []
        };
        const config = enrichCommunityServerConfig(server);
        expect(config.commandLineService?.command).toBe('npx -y @modelcontextprotocol/server-github');
        expect(config.configurationSchema).toContain('GITHUB_PERSONAL_ACCESS_TOKEN');
    });

    it('should inject GitHub schema for Forked GitHub server but keep custom command', () => {
        const server: CommunityServer = {
            name: 'My GitHub Fork',
            url: 'https://github.com/user/my-github-fork',
            description: 'Custom GitHub Integration',
            category: 'Dev',
            tags: []
        };
        const config = enrichCommunityServerConfig(server);
        // Should use the fork's repo name for command
        expect(config.commandLineService?.command).toBe('npx -y my-github-fork');
        // But should still detect it's a GitHub service and inject schema
        expect(config.configurationSchema).toContain('GITHUB_PERSONAL_ACCESS_TOKEN');
    });

    it('should inject Slack schema for Slack server', () => {
        const server: CommunityServer = {
            name: 'Slack',
            url: 'https://github.com/modelcontextprotocol/server-slack',
            description: 'Slack Integration',
            category: 'Communication',
            tags: []
        };
        const config = enrichCommunityServerConfig(server);
        expect(config.commandLineService?.command).toBe('npx -y @modelcontextprotocol/server-slack');
        expect(config.configurationSchema).toContain('SLACK_BOT_TOKEN');
        expect(config.configurationSchema).toContain('SLACK_TEAM_ID');
    });

    it('should use uvx for python servers (tag based)', () => {
        const server: CommunityServer = {
            name: 'Python Tool',
            url: 'https://github.com/user/repo',
            description: 'Python stuff',
            category: 'Tools',
            tags: ['🐍']
        };
        const config = enrichCommunityServerConfig(server);
        expect(config.commandLineService?.command).toBe('uvx repo');
    });

    it('should fallback gracefully for unknown servers', () => {
        const server: CommunityServer = {
            name: 'Unknown Thing',
            url: 'https://github.com/user/unknown-repo',
            description: 'Mystery',
            category: 'Tools',
            tags: []
        };
        const config = enrichCommunityServerConfig(server);
        expect(config.commandLineService?.command).toBe('npx -y unknown-repo');
        expect(config.configurationSchema).toBe("");
    });

    it('should handle AWS server', () => {
        const server: CommunityServer = {
            name: 'AWS',
            url: 'https://github.com/modelcontextprotocol/server-aws',
            description: 'AWS Integration',
            category: 'Cloud',
            tags: []
        };
        const config = enrichCommunityServerConfig(server);
        expect(config.configurationSchema).toContain('AWS_ACCESS_KEY_ID');
    });
});
