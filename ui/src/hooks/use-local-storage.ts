/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useCallback } from "react";

/**
 * useLocalStorage is a custom React hook for persisting state in the browser's LocalStorage.
 *
 * It syncs the state with LocalStorage so that the data persists across browser sessions.
 * It also listens for changes to the key in LocalStorage from other tabs/windows.
 *
 * @template T The type of the value to be stored.
 * @param key - The unique key used to store the value in LocalStorage.
 * @param initialValue - The initial value to use if the key does not exist in LocalStorage.
 * @returns A tuple containing:
 *          - storedValue: The current value.
 *          - setValue: A function to update the value (accepts a value or a function).
 *          - isInitialized: A boolean indicating if the value has been read from LocalStorage.
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
