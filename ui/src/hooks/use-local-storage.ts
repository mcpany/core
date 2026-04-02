/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useCallback } from "react";

/**
 * Summary: A hook that manages state synchronized with the browser's localStorage.
 *
 * Params:
 *   - key (string): The unique key under which the value is stored in localStorage.
 *   - initialValue (T): The initial value to use if no value is currently found in localStorage.
 *
 * Returns:
 *   - [T, (value: T | ((val: T) => T)) => void, boolean]: A tuple containing the current state value, a setter function to update it, and a boolean indicating if initialization from localStorage has completed.
 *
 * Errors:
 *   - Catches and logs errors to the console if JSON parsing or localStorage access fails.
 *
 * Side Effects:
 *   - Reads from and writes to the browser's `window.localStorage` synchronously.
 */
export function useLocalStorage<T>(key: string, initialValue: T): [T, (value: T | ((val: T) => T)) => void, boolean] {
  const [storedValue, setStoredValue] = useState<T>(initialValue);
  const [isInitialized, setIsInitialized] = useState(false);

  useEffect(() => {
    if (typeof window === "undefined") return;

    try {
      const item = window.localStorage.getItem(key);
      if (item) {
        setStoredValue(JSON.parse(item));
      }
    } catch (error) {
      console.error(error);
    } finally {
      setIsInitialized(true);
    }
  }, [key]);

  const setValue = useCallback((value: T | ((val: T) => T)) => {
    try {
      setStoredValue((prev) => {
        const valueToStore = value instanceof Function ? value(prev) : value;
        if (typeof window !== "undefined") {
            window.localStorage.setItem(key, JSON.stringify(valueToStore));
        }
        return valueToStore;
      });
    } catch (error) {
      console.error(error);
    }
  }, [key]);

  return [storedValue, setValue, isInitialized];
}
