/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from "react";

const HISTORY_KEY = "mcpany-tool-history";

/**
 * Interface representing a persisted tool execution entry.
 */
export interface HistoryEntry {
  /** Unique ID of the entry */
  id: string;
  /** ISO timestamp of when the tool was executed */
  timestamp: string;
  /** Name of the tool executed */
  toolName: string;
  /** Input arguments used for execution */
  args: any;
  /** Result or output of the tool execution */
  result: any;
  /** Duration of execution in milliseconds */
  duration: number;
  /** Status of the execution (success/error) */
  status: "success" | "error";
}

/**
 * Hook to manage local execution history of tool runs.
 * Persists data to localStorage.
 *
 * @returns The history state and functions to add/clear entries.
 */
export function useExecutionHistory() {
  const [history, setHistory] = useState<HistoryEntry[]>([]);
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    try {
      const item = window.localStorage.getItem(HISTORY_KEY);
      if (item) {
        setHistory(JSON.parse(item));
      }
    } catch (error) {
      console.error("Failed to load history from local storage", error);
    } finally {
      setIsLoaded(true);
    }
  }, []);

  useEffect(() => {
    if (!isLoaded) return;
    try {
      window.localStorage.setItem(HISTORY_KEY, JSON.stringify(history));
    } catch (error) {
      console.error("Failed to save history to local storage", error);
    }
  }, [history, isLoaded]);

  const addEntry = (entry: Omit<HistoryEntry, "id" | "timestamp">) => {
    const newEntry: HistoryEntry = {
      ...entry,
      id: crypto.randomUUID(),
      timestamp: new Date().toISOString(),
    };
    setHistory((prev) => [newEntry, ...prev].slice(0, 50)); // Keep last 50
  };

  const clearHistory = () => {
    setHistory([]);
  };

  return { history, addEntry, clearHistory, isLoaded };
}
