/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { marketplaceService, ServiceCollection } from './marketplace-service';

// Mock localStorage
const localStorageMock = (function() {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] || null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value.toString();
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key];
    }),
    clear: vi.fn(() => {
      store = {};
    }),
  };
})();

Object.defineProperty(window, 'localStorage', { value: localStorageMock });

describe('marketplaceService', () => {
  beforeEach(() => {
    localStorageMock.clear();
    vi.clearAllMocks();
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('fetchCommunityServers', () => {
    it('should parse valid markdown correctly', async () => {
      const mockMarkdown = `
# Awesome MCP Servers

## 📂 File Systems
* [FileSystem](https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem) 📁 - Access local files
* [Google Drive](https://github.com/modelcontextprotocol/servers/tree/main/src/gdrive) ☁️ - Access Google Drive files

## 🛠️ Developer Tools
- [GitHub](https://github.com/modelcontextprotocol/servers/tree/main/src/github) 🐙 - Access GitHub repositories
`;

      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        text: () => Promise.resolve(mockMarkdown),
      });

      const servers = await marketplaceService.fetchCommunityServers();

      expect(servers).toHaveLength(3);
      expect(servers[0]).toEqual({
        category: '📂 File Systems',
        name: 'FileSystem',
        url: 'https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem',
        description: 'Access local files',
        tags: ['📁']
      });
      expect(servers[1]).toEqual({
        category: '📂 File Systems',
        name: 'Google Drive',
        url: 'https://github.com/modelcontextprotocol/servers/tree/main/src/gdrive',
        description: 'Access Google Drive files',
        tags: ['☁️']
      });
      expect(servers[2]).toEqual({
        category: '🛠️ Developer Tools',
        name: 'GitHub',
        url: 'https://github.com/modelcontextprotocol/servers/tree/main/src/github',
        description: 'Access GitHub repositories',
        tags: ['🐙']
      });
    });

    it('should handle fetch errors gracefully', async () => {
      global.fetch = vi.fn().mockRejectedValue(new Error('Network error'));
      const servers = await marketplaceService.fetchCommunityServers();
      expect(servers).toEqual([]);
    });

    it('should handle non-ok responses', async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: false,
      });
      const servers = await marketplaceService.fetchCommunityServers();
      expect(servers).toEqual([]);
    });

    it('should handle empty markdown', async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        text: () => Promise.resolve(''),
      });
      const servers = await marketplaceService.fetchCommunityServers();
      expect(servers).toEqual([]);
    });
  });

  describe('fetchExternalServers', () => {
    it('should return linear server for mcpmarket', async () => {
      const servers = await marketplaceService.fetchExternalServers('mcpmarket');
      expect(servers).toHaveLength(1);
      expect(servers[0].id).toBe('linear');
      expect(servers[0].name).toBe('Linear');
    });

    it('should return empty array for unknown marketplace', async () => {
      const servers = await marketplaceService.fetchExternalServers('unknown');
      expect(servers).toEqual([]);
    });
  });

  describe('Local Collections (localStorage)', () => {
    it('should save and fetch collections', () => {
      const collection: ServiceCollection = {
        name: 'My Collection',
        description: 'Test Description',
        author: 'Me',
        version: '1.0.0',
        services: []
      };

      marketplaceService.saveLocalCollection(collection);
      expect(localStorageMock.setItem).toHaveBeenCalledWith('mcp_local_collections', expect.any(String));

      const stored = marketplaceService.fetchLocalCollections();
      expect(stored).toHaveLength(1);
      expect(stored[0]).toEqual(collection);
    });

    it('should update existing collection by name', () => {
      const collection1: ServiceCollection = {
        name: 'My Collection',
        description: 'ver 1',
        author: 'Me',
        version: '1.0.0',
        services: []
      };
      marketplaceService.saveLocalCollection(collection1);

      const collection2: ServiceCollection = {
        name: 'My Collection', // Same name
        description: 'ver 2',
        author: 'Me',
        version: '1.0.1',
        services: []
      };
      marketplaceService.saveLocalCollection(collection2);

      const stored = marketplaceService.fetchLocalCollections();
      expect(stored).toHaveLength(1);
      expect(stored[0].description).toBe('ver 2');
    });

    it('should delete a collection', () => {
      const collection: ServiceCollection = {
        name: 'My Collection',
        description: 'Test Description',
        author: 'Me',
        version: '1.0.0',
        services: []
      };
      marketplaceService.saveLocalCollection(collection);
      marketplaceService.deleteLocalCollection('My Collection');

      const stored = marketplaceService.fetchLocalCollections();
      expect(stored).toEqual([]);
    });

    it('should handle corrupted local storage', () => {
      localStorageMock.getItem.mockReturnValue('invalid json');
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      const stored = marketplaceService.fetchLocalCollections();
      expect(stored).toEqual([]);

      consoleSpy.mockRestore();
    });
  });

  describe('importCollection', () => {
    it('should return a mock collection', async () => {
      const collection = await marketplaceService.importCollection('http://example.com/collection.json');
      expect(collection.name).toBe('Imported Collection');
      expect(collection.description).toContain('http://example.com/collection.json');
    });
  });
});
