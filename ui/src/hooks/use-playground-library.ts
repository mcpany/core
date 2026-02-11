/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect, useCallback } from "react";
import { apiClient, RequestCollection, SavedRequest } from "@/lib/client";
import { toast } from "sonner";

export function usePlaygroundLibrary() {
  const [collections, setCollections] = useState<RequestCollection[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchCollections = useCallback(async () => {
    setLoading(true);
    try {
        const data = await apiClient.listRequestCollections();
        setCollections(data);
    } catch (e) {
        console.error(e);
        toast.error("Failed to load collections");
    } finally {
        setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCollections();
  }, [fetchCollections]);

  const createCollection = useCallback(async (name: string, description?: string) => {
    try {
        const newCol: RequestCollection = {
            id: crypto.randomUUID(),
            name,
            description: description || "",
            requests: [],
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
        };
        const saved = await apiClient.saveRequestCollection(newCol);
        setCollections(prev => [...prev, saved]);
        return saved;
    } catch (e) {
        toast.error("Failed to create collection");
        throw e;
    }
  }, []);

  const deleteCollection = useCallback(async (id: string) => {
    try {
        await apiClient.deleteRequestCollection(id);
        setCollections(prev => prev.filter(c => c.id !== id));
    } catch (e) {
        toast.error("Failed to delete collection");
    }
  }, []);

  const saveRequestToCollection = useCallback(async (collectionId: string, request: Omit<SavedRequest, "id" | "createdAt">) => {
    try {
        const collection = collections.find(c => c.id === collectionId);
        if (!collection) throw new Error("Collection not found");

        const newRequest: SavedRequest = {
            ...request,
            id: crypto.randomUUID(),
            createdAt: new Date().toISOString(),
        };

        const updatedCollection = {
            ...collection,
            requests: [...collection.requests, newRequest],
            updatedAt: new Date().toISOString(),
        };

        const saved = await apiClient.saveRequestCollection(updatedCollection);
        setCollections(prev => prev.map(c => c.id === collectionId ? saved : c));
    } catch (e) {
        toast.error("Failed to save request");
        console.error(e);
    }
  }, [collections]);

  const deleteRequestFromCollection = useCallback(async (collectionId: string, requestId: string) => {
    try {
        const collection = collections.find(c => c.id === collectionId);
        if (!collection) return;

        const updatedCollection = {
            ...collection,
            requests: collection.requests.filter(r => r.id !== requestId),
            updatedAt: new Date().toISOString(),
        };

        const saved = await apiClient.saveRequestCollection(updatedCollection);
        setCollections(prev => prev.map(c => c.id === collectionId ? saved : c));
    } catch (e) {
        toast.error("Failed to delete request");
    }
  }, [collections]);

  return {
    collections,
    createCollection,
    deleteCollection,
    saveRequestToCollection,
    deleteRequestFromCollection,
    loading
  };
}
