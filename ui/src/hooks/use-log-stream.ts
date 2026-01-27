/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useRef } from "react";

export type LogLevel = "INFO" | "WARN" | "ERROR" | "DEBUG";

export interface LogEntry {
  id: string;
  timestamp: string;
  level: LogLevel;
  message: string;
  source?: string;
  metadata?: Record<string, unknown>;
  searchStr?: string;
  formattedTime?: string;
}

const timeFormatter = typeof Intl !== 'undefined' ? new Intl.DateTimeFormat(undefined, {
  hour: 'numeric',
  minute: 'numeric',
  second: 'numeric',
}) : null;

interface UseLogStreamOptions {
  sourceFilter?: string; // Optional hard filter for source (e.g. for specific service diagnostics)
}

export function useLogStream(options?: UseLogStreamOptions) {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const logBufferRef = useRef<LogEntry[]>([]);
  const isPausedRef = useRef(isPaused);

  useEffect(() => {
    isPausedRef.current = isPaused;
  }, [isPaused]);

  useEffect(() => {
    const flushInterval = setInterval(() => {
      if (logBufferRef.current.length > 0) {
        setLogs((prev) => {
          const buffer = logBufferRef.current;
          logBufferRef.current = [];
          const MAX_LOGS = 1000;

          if (prev.length + buffer.length <= MAX_LOGS) {
            return [...prev, ...buffer];
          }
          if (buffer.length >= MAX_LOGS) {
            return buffer.slice(buffer.length - MAX_LOGS);
          }
          const keepCount = MAX_LOGS - buffer.length;
          return [...prev.slice(-keepCount), ...buffer];
        });
      }
    }, 100);

    const connect = () => {
      // Determine protocol (ws or wss)
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.host;
      const wsUrl = `${protocol}//${host}/api/v1/ws/logs`;
      const ws = new WebSocket(wsUrl);

      ws.onopen = () => setIsConnected(true);

      ws.onmessage = (event) => {
        if (isPausedRef.current) return;
        if (document.hidden) return;

        try {
          const newLog: LogEntry = JSON.parse(event.data);

          // Client-side hard filtering
          if (options?.sourceFilter && options.sourceFilter !== "ALL") {
             let match = false;
             if (newLog.source === options.sourceFilter) match = true;
             else if (newLog.metadata?.component === options.sourceFilter) {
                 match = true;
                 if (!newLog.source) newLog.source = options.sourceFilter;
             }

             if (!match) return;
          }

          newLog.searchStr = (newLog.message + " " + (newLog.source || "")).toLowerCase();
          newLog.formattedTime = timeFormatter
            ? timeFormatter.format(new Date(newLog.timestamp))
            : new Date(newLog.timestamp).toLocaleTimeString();

          logBufferRef.current.push(newLog);
        } catch (e) {
          console.error("Failed to parse log message", e);
        }
      };

      ws.onclose = () => {
        setIsConnected(false);
        setTimeout(connect, 3000);
      };

      ws.onerror = (err) => {
        console.error("WebSocket error", err);
        ws.close();
      };

      wsRef.current = ws;
    };

    connect();

    return () => {
      wsRef.current?.close();
      clearInterval(flushInterval);
    };
  }, [options?.sourceFilter]);

  const clearLogs = () => setLogs([]);

  return {
    logs,
    isConnected,
    isPaused,
    setIsPaused,
    clearLogs,
  };
}
