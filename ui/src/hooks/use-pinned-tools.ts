/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from "react";

const STORAGE_KEY = "mcpany-pinned-tools";

/**
 * Hook to manage the list of pinned tools in the user interface.
 *
 * It persists the list of pinned tool names in localStorage so they remain pinned across sessions.
 *
 * @returns An object containing:
 * - `pinnedTools`: The array of pinned tool names.
 * - `togglePin`: A function to toggle the pinned state of a tool.
 * - `isPinned`: A function to check if a specific tool is pinned.
 * - `isLoaded`: A boolean indicating if the initial load from localStorage is complete.
 */
export function usePinnedTools() {
  const [pinnedTools, setPinnedTools] = useState<string[]>([]);
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    try {
      const item = window.localStorage.getItem(STORAGE_KEY);
      if (item) {
        setPinnedTools(JSON.parse(item));
      }
    } catch (error) {
      console.error("Failed to load pinned tools from local storage", error);
    } finally {
      setIsLoaded(true);
    }
  }, []);

  const togglePin = (toolName: string) => {
    setPinnedTools((prev) => {
      const newPinned = prev.includes(toolName)
        ? prev.filter((t) => t !== toolName)
        : [...prev, toolName];

      try {
        window.localStorage.setItem(STORAGE_KEY, JSON.stringify(newPinned));
      } catch (error) {
        console.error("Failed to save pinned tools to local storage", error);
      }

      return newPinned;
    });
  };

  const isPinned = (toolName: string) => {
    return pinnedTools.includes(toolName);
  };

  return { pinnedTools, togglePin, isPinned, isLoaded };
}
