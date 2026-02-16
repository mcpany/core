/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useCallback } from "react";

/**
 * Manages state synchronized with localStorage.
 *
 * @remarks
 * This hook persists the state to the browser's localStorage, ensuring it survives page reloads.
 * It initializes the state from localStorage if available, otherwise uses the provided initial value.
 *
 * @template T - The type of the value to store.
 *
 * @param key - string. The unique key under which the value is stored in localStorage.
 * @param initialValue - T. The initial value to use if no value is found in localStorage.
 *
 * @returns [T, (value: T | ((val: T) => T)) => void, boolean] - A tuple containing:
 *          1. The current value.
 *          2. A setter function to update the value.
 *          3. A boolean indicating if the value has been initialized from storage.
 *
 * @sideeffects
 * - Reads from and writes to `window.localStorage`.
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
