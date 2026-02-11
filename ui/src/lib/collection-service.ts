/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

/**
 * Represents a single test case within a collection.
 */
export interface TestCase {
  id: string;
  name: string;
  toolName: string;
  args: Record<string, unknown>;
  createdAt: number;
}

/**
 * Represents a collection of test cases.
 */
export interface Collection {
  id: string;
  name: string;
  description?: string;
  items: TestCase[];
  createdAt: number;
}

const STORAGE_KEY = "mcpany-collections";

/**
 * Service for managing Playground Collections in localStorage.
 */
export const CollectionService = {
  /**
   * List all collections.
   * Includes seeding logic for first run.
   */
  list: (): Collection[] => {
    if (typeof window === "undefined") return [];

    const stored = localStorage.getItem(STORAGE_KEY);
    if (!stored) {
      // Seed with demo data if empty
      const demoCollection: Collection = {
        id: "demo-collection",
        name: "My First Collection",
        description: "A starter collection to organize your tool tests.",
        createdAt: Date.now(),
        items: []
      };
      localStorage.setItem(STORAGE_KEY, JSON.stringify([demoCollection]));
      return [demoCollection];
    }

    try {
      return JSON.parse(stored);
    } catch (e) {
      console.error("Failed to parse collections", e);
      return [];
    }
  },

  /**
   * Get a specific collection by ID.
   */
  get: (id: string): Collection | undefined => {
    const collections = CollectionService.list();
    return collections.find(c => c.id === id);
  },

  /**
   * Save (Create or Update) a collection.
   */
  save: (collection: Collection): void => {
    const collections = CollectionService.list();
    const index = collections.findIndex(c => c.id === collection.id);

    if (index >= 0) {
      collections[index] = collection;
    } else {
      collections.push(collection);
    }

    localStorage.setItem(STORAGE_KEY, JSON.stringify(collections));
  },

  /**
   * Delete a collection.
   */
  delete: (id: string): void => {
    const collections = CollectionService.list();
    const filtered = collections.filter(c => c.id !== id);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(filtered));
  },

  /**
   * Add a test case to a collection.
   */
  addTestCase: (collectionId: string, testCase: TestCase): void => {
    const collections = CollectionService.list();
    const collection = collections.find(c => c.id === collectionId);
    if (collection) {
      collection.items.push(testCase);
      localStorage.setItem(STORAGE_KEY, JSON.stringify(collections));
    }
  },

  /**
   * Remove a test case from a collection.
   */
  removeTestCase: (collectionId: string, testCaseId: string): void => {
    const collections = CollectionService.list();
    const collection = collections.find(c => c.id === collectionId);
    if (collection) {
      collection.items = collection.items.filter(item => item.id !== testCaseId);
      localStorage.setItem(STORAGE_KEY, JSON.stringify(collections));
    }
  }
};
