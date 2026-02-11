/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { CollectionService, Collection, TestCase } from './collection-service';

describe('CollectionService', () => {
  beforeEach(() => {
    // Mock localStorage
    const store: Record<string, string> = {};
    const localStorageMock = {
      getItem: vi.fn((key: string) => store[key] || null),
      setItem: vi.fn((key: string, value: string) => {
        store[key] = value.toString();
      }),
      clear: vi.fn(() => {
        for (const key in store) delete store[key];
      }),
      removeItem: vi.fn((key: string) => {
        delete store[key];
      }),
      key: vi.fn(),
      length: 0,
    };
    Object.defineProperty(window, 'localStorage', { value: localStorageMock, writable: true });
  });

  it('should seed data if empty', () => {
    const collections = CollectionService.list();
    expect(collections.length).toBe(1);
    expect(collections[0].id).toBe('demo-collection');
    expect(localStorage.getItem('mcpany-collections')).toBeTruthy();
  });

  it('should save and retrieve collections', () => {
    // Clear seed
    localStorage.setItem('mcpany-collections', '[]');

    const collection: Collection = {
      id: '1',
      name: 'Test Collection',
      items: [],
      createdAt: 123
    };

    CollectionService.save(collection);

    const listed = CollectionService.list();
    expect(listed).toHaveLength(1);
    expect(listed[0].name).toBe('Test Collection');

    const retrieved = CollectionService.get('1');
    expect(retrieved).toEqual(collection);
  });

  it('should add a test case', () => {
    // Seed
    CollectionService.list(); // triggers seed

    const testCase: TestCase = {
      id: 'tc1',
      name: 'Case 1',
      toolName: 'echo',
      args: { msg: 'hi' },
      createdAt: 123
    };

    CollectionService.addTestCase('demo-collection', testCase);

    const col = CollectionService.get('demo-collection');
    expect(col?.items).toHaveLength(1);
    expect(col?.items[0].name).toBe('Case 1');
  });

  it('should remove a test case', () => {
    // Seed
    CollectionService.list();
    const testCase: TestCase = { id: 'tc1', name: 'Case 1', toolName: 'echo', args: {}, createdAt: 123 };
    CollectionService.addTestCase('demo-collection', testCase);

    CollectionService.removeTestCase('demo-collection', 'tc1');

    const col = CollectionService.get('demo-collection');
    expect(col?.items).toHaveLength(0);
  });
});
