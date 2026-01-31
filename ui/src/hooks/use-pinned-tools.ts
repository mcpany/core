/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from "react";

const STORAGE_KEY = "mcpany-pinned-tools";

/**
 * Hook for pinnedtools.
 * @returns The result.
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

  const bulkSetPins = (names: string[], shouldPin: boolean) => {
    setPinnedTools((prev) => {
      const currentSet = new Set(prev);
      names.forEach((name) => {
        if (shouldPin) currentSet.add(name);
        else currentSet.delete(name);
      });
      const newPinned = Array.from(currentSet);

      try {
        window.localStorage.setItem(STORAGE_KEY, JSON.stringify(newPinned));
      } catch (error) {
        console.error("Failed to save pinned tools", error);
      }

      return newPinned;
    });
  };

  const isPinned = (toolName: string) => {
    return pinnedTools.includes(toolName);
  };

  return { pinnedTools, togglePin, bulkSetPins, isPinned, isLoaded };
}
