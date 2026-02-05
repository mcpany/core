/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useCallback } from "react";

/**
 * Persists a stateful value in `localStorage`, similar to `useState`.
 *
 * This hook synchronizes the state with the browser's `localStorage` to persist data
 * across page reloads. It handles JSON serialization and deserialization automatically.
 *
 * @param key - The key used to store the value in `localStorage`.
 * @param initialValue - The initial value to use if no value is found in storage.
 * @returns A tuple containing the current value, a setter function, and an initialization status boolean.
 *
 * @throws {Error} Logs an error to the console if `JSON.parse` or `localStorage.setItem` fails, but does not throw.
 *
 * @example
 * ```tsx
 * const [theme, setTheme, isLoaded] = useLocalStorage("theme", "light");
 * ```
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
