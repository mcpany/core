/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useEffect, useRef } from 'react';

/**
 * usePolling is a hook to poll a callback at a specified interval.
 * It automatically stops polling when the document is hidden and resumes when visible.
 *
 * @param callback - The function to call.
 * @param delay - The interval in milliseconds. If null, polling is paused.
 */
export function usePolling(callback: () => void, delay: number | null) {
  const savedCallback = useRef(callback);

  // Remember the latest callback
  useEffect(() => {
    savedCallback.current = callback;
  }, [callback]);

  // Set up the interval and visibility listener
  useEffect(() => {
    if (delay === null) return;

    let id: NodeJS.Timeout | null = null;

    const tick = () => {
      if (savedCallback.current) {
        savedCallback.current();
      }
    };

    const handleVisibilityChange = () => {
      if (document.hidden) {
        // Clear interval when hidden
        if (id) {
          clearInterval(id);
          id = null;
        }
      } else {
        // Resume immediately when visible
        if (!id) {
          tick(); // Execute immediately on resume
          id = setInterval(tick, delay);
        }
      }
    };

    // Initial setup
    if (!document.hidden) {
      id = setInterval(tick, delay);
    }

    document.addEventListener('visibilitychange', handleVisibilityChange);

    return () => {
      if (id) clearInterval(id);
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, [delay]);
}
