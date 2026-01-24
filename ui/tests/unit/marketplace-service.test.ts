/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi } from 'vitest';
import { marketplaceService } from '../../src/lib/marketplace-service';

// Mock fetch global
global.fetch = vi.fn();

// Helper to mock fetch implementation
const mockFetch = (response: Partial<Response>) => {
    (global.fetch as unknown as { mockResolvedValue: (val: unknown) => void }).mockResolvedValue(response);
};

describe('marketplaceService', () => {
  describe('fetchCommunityServers', () => {
    it('should parse the Awesome list markdown correctly', async () => {
      const mockMarkdown = `
# Awesome MCP Servers

## 📂 File Systems

- [mcp-server-filesystem](https://github.com/modelcontextprotocol/server-filesystem) 📇 🏠 - Direct local file system access.
- [box/mcp-server-box-remote](https://github.com/box/mcp-server-box-remote) 🎖️ ☁️ - The Box MCP server allows...

## ☁️ Cloud Platforms

- [cloudflare/mcp-server-cloudflare](https://github.com/cloudflare/mcp-server-cloudflare) 🎖️ 📇 ☁️ - Integration with Cloudflare...
      `;

      mockFetch({
        ok: true,
        text: async () => mockMarkdown,
      } as Response);

      const servers = await marketplaceService.fetchCommunityServers();

      expect(servers).toHaveLength(3);

      expect(servers[0]).toEqual({
        category: '📂 File Systems',
        name: 'mcp-server-filesystem',
        url: 'https://github.com/modelcontextprotocol/server-filesystem',
        description: 'Direct local file system access.',
        tags: ['📇', '🏠']
      });

      expect(servers[1]).toEqual({
        category: '📂 File Systems',
        name: 'box/mcp-server-box-remote',
        url: 'https://github.com/box/mcp-server-box-remote',
        description: 'The Box MCP server allows...',
        tags: ['🎖️', '☁️']
      });

      expect(servers[2]).toEqual({
        category: '☁️ Cloud Platforms',
        name: 'cloudflare/mcp-server-cloudflare',
        url: 'https://github.com/cloudflare/mcp-server-cloudflare',
        description: 'Integration with Cloudflare...',
        tags: ['🎖️', '📇', '☁️']
      });
    });

    it('should handle fetch errors gracefully', async () => {
        mockFetch({
            ok: false,
        } as Response);

        const servers = await marketplaceService.fetchCommunityServers();
        expect(servers).toEqual([]);
    });
  });
});
