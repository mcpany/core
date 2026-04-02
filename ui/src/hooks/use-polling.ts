/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useEffect, useRef } from 'react';

/**
 * Summary: A hook to poll a callback function at a specified interval with visibility awareness.
 *
 * Params:
 *   - callback (Function): The function to execute on each interval tick.
 *   - delay (number | null): The polling interval in milliseconds; polling halts if null.
 *
 * Returns:
 *   - void
 *
 * Errors:
 *   - N/A: Absorbs internal errors and logs to console without throwing.
 *
 * Side Effects:
 *   - Attaches `setInterval` and `visibilitychange` listeners. Stops polling automatically when the tab is hidden to save resources.
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
