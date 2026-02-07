/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useCallback } from "react";

/**
 * A custom hook to persist state in localStorage.
 * Synchronizes the state with the browser's localStorage, allowing data to persist across page reloads.
 *
 * @param key - The key under which the value is stored in localStorage.
 * @param initialValue - The initial value to use if no value is found in localStorage.
 * @returns A tuple containing:
 *  - storedValue: The current value.
 *  - setValue: A function to update the value.
 *  - isInitialized: A boolean indicating if the value has been read from localStorage.
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
