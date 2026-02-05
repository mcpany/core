/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from "react";

const STORAGE_KEY = "mcpany-recent-tools";
const MAX_RECENT_TOOLS = 5;

/**
 * Manages the history of recently used tools.
 *
 * This hook persists a list of recently used tool names in `localStorage`, maintaining a maximum size of 5.
 * It provides functions to add new usage (which moves the tool to the top) and clear history.
 *
 * @returns An object containing the recent tools list and management functions.
 * - `recentTools`: Array of tool names, most recent first.
 * - `addRecent`: Function to record a tool usage. Moves tool to front.
 * - `removeRecent`: Function to remove a tool from history.
 * - `clearRecent`: Function to clear all history.
 * - `isLoaded`: Boolean indicating if data has been loaded from storage.
 *
 * @remarks
 * Side Effects:
 * - Reads from `window.localStorage` on mount.
 * - Writes to `window.localStorage` whenever the list changes.
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
