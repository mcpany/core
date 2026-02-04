/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from "react";

const STORAGE_KEY = "mcpany-recent-tools";
const MAX_RECENT_TOOLS = 5;

/**
 * A hook that manages and persists a list of recently used tool names.
 *
 * It stores the list in localStorage and limits it to `MAX_RECENT_TOOLS`.
 *
 * @returns An object containing the list of recent tools and functions to add, remove, or clear them.
 */
export function useRecentTools() {
  const [recentTools, setRecentTools] = useState<string[]>([]);
  const [isLoaded, setIsLoaded] = useState(false);

  // Load from storage on mount
  useEffect(() => {
    try {
      const item = window.localStorage.getItem(STORAGE_KEY);
      if (item) {
        setRecentTools(JSON.parse(item));
      }
    } catch (error) {
      console.error("Failed to load recent tools from local storage", error);
    } finally {
      setIsLoaded(true);
    }
  }, []);

  // Sync to storage on change
  useEffect(() => {
    if (!isLoaded) return;

    try {
      if (recentTools.length === 0) {
        window.localStorage.removeItem(STORAGE_KEY);
      } else {
        window.localStorage.setItem(STORAGE_KEY, JSON.stringify(recentTools));
      }
    } catch (error) {
      console.error("Failed to save recent tools to local storage", error);
    }
  }, [recentTools, isLoaded]);

  const addRecent = (toolName: string) => {
    setRecentTools((prev) => {
      // Remove if exists, then add to front
      const filtered = prev.filter((t) => t !== toolName);
      return [toolName, ...filtered].slice(0, MAX_RECENT_TOOLS);
    });
  };

  const removeRecent = (toolName: string) => {
    setRecentTools((prev) => prev.filter((t) => t !== toolName));
  };

  const clearRecent = () => {
    setRecentTools([]);
  };

  return { recentTools, addRecent, removeRecent, clearRecent, isLoaded };
}
