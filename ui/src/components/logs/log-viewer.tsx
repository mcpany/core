/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */
import * as React from "react"
import { lazy, Suspense } from "react"
import {
  ChevronRight,
  ChevronDown
} from "lucide-react"
import { cn } from "@/lib/utils"

// Lazy load Virtuoso to avoid SSR issues
const Virtuoso = lazy(() => import('react-virtuoso').then((m) => ({ default: m.Virtuoso })));

// Lazy load the syntax highlighter
const JsonViewer = lazy(() => import('./json-viewer'));

/**
 * LogLevel defines the severity of a log entry.
 */
export type LogLevel = "INFO" | "WARN" | "ERROR" | "DEBUG"

/**
 * LogEntry represents a single structured log message.
 */
export interface LogEntry {
  id: string
  timestamp: string
  level: LogLevel
  message: string
  source?: string
  metadata?: Record<string, unknown>
  formattedTime?: string
}

interface LogViewerProps {
  logs: LogEntry[];
  highlightRegex: RegExp | null;
  isPaused?: boolean;
}

// Optimization: Reuse DateTimeFormat instance
/**
 * timeFormatter is a shared Intl.DateTimeFormat instance for formatting log timestamps.
 */
export const timeFormatter = typeof Intl !== 'undefined' ? new Intl.DateTimeFormat(undefined, {
  hour: 'numeric',
  minute: 'numeric',
  second: 'numeric',
}) : null;

<<<<<<< HEAD
const getLevelBadgeClasses = (level: LogLevel) => {
  switch (level) {
    case "INFO": return "bg-blue-500/10 text-blue-500 border border-blue-500/20 shadow-[0_0_10px_rgba(59,130,246,0.1)]"
    case "WARN": return "bg-yellow-500/10 text-yellow-500 border border-yellow-500/20 shadow-[0_0_10px_rgba(234,179,8,0.1)]"
    case "ERROR": return "bg-red-500/10 text-red-500 border border-red-500/30 shadow-[0_0_15px_rgba(239,68,68,0.2)]"
    case "DEBUG": return "bg-gray-500/10 text-gray-400 border border-gray-500/20"
    default: return "bg-background text-foreground border border-border"
=======
const getLevelColor = (level: LogLevel) => {
  switch (level) {
    case "INFO": return "text-blue-400"
    case "WARN": return "text-yellow-400"
    case "ERROR": return "text-red-400"
    case "DEBUG": return "text-gray-400"
    default: return "text-foreground"
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
  }
}

const getSourceHue = (source: string) => {
  let hash = 0;
  for (let i = 0; i < source.length; i++) {
    hash = source.charCodeAt(i) + ((hash << 5) - hash);
  }
  return Math.abs(hash % 360);
};

const isLikelyJson = (str: string): boolean => {
  if (typeof str !== 'string') return false;
  const trimmed = str.trim();
  return (trimmed.startsWith('{') && trimmed.endsWith('}')) ||
         (trimmed.startsWith('[') && trimmed.endsWith(']'));
};

const safeParseJson = (str: string): unknown | null => {
  if (typeof str !== 'string') return null;
  try {
    return JSON.parse(str);
  } catch {
    return null;
  }
};

/**
 * HighlightText component highlights portions of text that match a regular expression.
 * @param props - The component props.
 * @param props.text - The text to be evaluated and rendered.
 * @param props.regex - The regular expression used to find matches.
 * @returns The rendered highlighted text.
 */
const HighlightText = React.memo(({ text, regex }: { text: string; regex: RegExp | null }) => {
  if (!regex || !text) return <>{text}</>;

  const parts = text.split(regex);

  return (
    <>
      {parts.map((part, i) =>
        i % 2 === 1 ? (
          <mark key={i} className="bg-yellow-500/40 text-inherit rounded-sm px-0.5 -mx-0.5">
            {part}
          </mark>
        ) : (
          part
        )
      )}
    </>
  );
});
HighlightText.displayName = 'HighlightText';

/**
 * LogRow component represents a single row in the log viewer table.
 * @param props - The component props.
 * @param props.log - The log entry data to display.
 * @param props.highlightRegex - The regular expression for highlighting terms.
 * @returns The rendered row.
 */
const LogRow = React.memo(({ log, highlightRegex }: { log: LogEntry; highlightRegex: RegExp | null }) => {
  const duration = log.metadata?.duration as string | undefined
  const [isExpanded, setIsExpanded] = React.useState(false);

  const isPotentialJson = React.useMemo(() => isLikelyJson(log.message), [log.message]);

  const jsonContent = React.useMemo(() => {
    if (isExpanded && isPotentialJson) {
      return safeParseJson(log.message);
    }
    return null;
  }, [isExpanded, isPotentialJson, log.message]);

<<<<<<< HEAD
  const rowClasses = cn(
    "group flex flex-col items-start p-3 sm:p-4 rounded-xl transition-all duration-200 ease-out break-words border mb-2",
    "bg-background/40 backdrop-blur-md hover:bg-background/60 shadow-sm hover:shadow-md",
    log.level === "ERROR" ? "border-red-500/30" : "border-border/50",
    isExpanded && isPotentialJson ? "ring-1 ring-primary/20" : ""
  );

  return (
    <div
      className={rowClasses}
      style={{ contentVisibility: 'auto', containIntrinsicSize: '0 64px' } as React.CSSProperties}
    >
      <div
        className={cn("flex flex-row w-full items-start gap-3 sm:gap-4", isPotentialJson ? "cursor-pointer" : "")}
        onClick={() => isPotentialJson && setIsExpanded(!isExpanded)}
      >
          <div className="flex flex-col sm:flex-row items-start sm:items-center gap-2 sm:gap-4 shrink-0 min-w-[140px] pt-0.5">
              <span className={cn("inline-flex items-center justify-center px-2 py-0.5 rounded-md text-[10px] font-bold tracking-wider uppercase", getLevelBadgeClasses(log.level))}>
                {log.level}
              </span>
              <span className="text-muted-foreground whitespace-nowrap opacity-60 font-mono text-[10px] sm:text-xs hidden sm:inline-block">
                {log.formattedTime || (timeFormatter ? timeFormatter.format(new Date(log.timestamp)) : new Date(log.timestamp).toLocaleTimeString())}
              </span>
          </div>

          <div className="flex-1 min-w-0 flex flex-col">
            <div className="flex items-start justify-between gap-4">
              <div className="flex flex-col sm:flex-row sm:items-baseline gap-1 sm:gap-3 flex-1 min-w-0">
                {log.source && (
                  <span
                    className="inline-block truncate font-mono text-[10px] sm:text-xs shrink-0 text-[hsl(var(--source-hue),60%,40%)] dark:text-[hsl(var(--source-hue),60%,70%)] font-medium bg-[hsl(var(--source-hue),60%,40%,0.1)] dark:bg-[hsl(var(--source-hue),60%,70%,0.1)] px-1.5 py-0.5 rounded-sm"
                    style={{ "--source-hue": getSourceHue(log.source) } as React.CSSProperties}
                    title={log.source}
                  >
                    <HighlightText text={log.source} regex={highlightRegex} />
                  </span>
                )}
                <span className="text-foreground/90 font-mono text-[11px] sm:text-sm pl-0 flex-1 break-all whitespace-pre-wrap leading-relaxed">
                   <HighlightText text={log.message} regex={highlightRegex} />
                </span>
              </div>

              <div className="flex items-center gap-2 shrink-0 pt-0.5">
                 {duration && (
                  <span className="inline-flex items-center rounded bg-muted/50 border border-border/50 px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground font-mono">
                    {duration}
                  </span>
                )}
                 {isPotentialJson && (
                    <button
                      onClick={(e) => { e.stopPropagation(); setIsExpanded(!isExpanded); }}
                      className="text-muted-foreground hover:text-foreground transition-colors p-1 rounded-full hover:bg-muted"
                      aria-label={isExpanded ? "Collapse JSON" : "Expand JSON"}
                    >
                      {isExpanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                    </button>
                 )}
              </div>
            </div>

            <div className={cn("grid transition-all duration-300 ease-in-out", isExpanded && isPotentialJson ? "grid-rows-[1fr] opacity-100 mt-3" : "grid-rows-[0fr] opacity-0")}>
              <div className="overflow-hidden">
                {isExpanded && isPotentialJson && (
                  <div className="w-full max-w-full overflow-x-auto text-xs bg-[#1e1e1e] dark:bg-black/40 rounded-lg border border-border/40 p-4 shadow-inner">
                    {jsonContent ? (
                      <Suspense fallback={<div className="animate-pulse h-8 bg-muted/20 rounded" />}><JsonViewer data={jsonContent} /></Suspense>
                    ) : (
                      <div className="p-3 bg-red-500/10 text-red-400 rounded-md border border-red-500/20 italic text-sm">
                        Failed to parse JSON payload
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
=======
  return (
    <div
      className="group flex flex-col items-start hover:bg-white/5 p-2 sm:p-1 rounded transition-colors break-words border-b border-white/5 sm:border-0"
      style={{ contentVisibility: 'auto', containIntrinsicSize: '0 32px' } as React.CSSProperties}
    >
      <div className="flex flex-row w-full items-start gap-1 sm:gap-3">
          <div className="flex items-center gap-2 sm:contents">
              <span className="text-muted-foreground whitespace-nowrap opacity-50 text-[10px] sm:text-xs sm:mt-0.5">
                {log.formattedTime || (timeFormatter ? timeFormatter.format(new Date(log.timestamp)) : new Date(log.timestamp).toLocaleTimeString())}
              </span>
              <span className={cn("font-bold w-12 text-[10px] sm:text-xs sm:mt-0.5", getLevelColor(log.level))}>
                {log.level}
              </span>
              {log.source && (
                <span
                  className="sm:hidden inline-block truncate text-[10px] flex-1 text-right text-[hsl(var(--source-hue),60%,40%)] dark:text-[hsl(var(--source-hue),60%,70%)]"
                  style={{ "--source-hue": getSourceHue(log.source) } as React.CSSProperties}
                  title={log.source}
                >
                  [<HighlightText text={log.source} regex={highlightRegex} />]
                </span>
              )}
          </div>

          {log.source && (
            <span
              className="hidden sm:inline-block w-24 truncate text-xs mt-0.5 shrink-0 text-[hsl(var(--source-hue),60%,40%)] dark:text-[hsl(var(--source-hue),60%,70%)]"
              style={{ "--source-hue": getSourceHue(log.source) } as React.CSSProperties}
              title={log.source}
            >
              [<HighlightText text={log.source} regex={highlightRegex} />]
            </span>
          )}

          <div className="flex-1 min-w-0 flex flex-col">
            <span className="text-gray-300 text-xs sm:text-sm pl-0 flex items-start">
               {isPotentialJson && (
                  <button
                    onClick={() => setIsExpanded(!isExpanded)}
                    className="mr-1 mt-0.5 text-muted-foreground hover:text-foreground"
                    aria-label={isExpanded ? "Collapse JSON" : "Expand JSON"}
                  >
                    {isExpanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                  </button>
               )}
               <span className="break-all whitespace-pre-wrap">
                 <HighlightText text={log.message} regex={highlightRegex} />
               </span>
               {duration && (
                <span className="ml-2 inline-flex items-center rounded-sm bg-white/10 px-1.5 py-0.5 text-[10px] font-medium text-gray-400 font-mono shrink-0">
                  {duration}
                </span>
              )}
            </span>

            {isExpanded && isPotentialJson && (
              <div className="mt-2 w-full max-w-full overflow-hidden text-xs">
                {jsonContent ? (
                  <Suspense fallback={null}><JsonViewer data={jsonContent} /></Suspense>
                ) : (
                  <div className="p-2 bg-muted/20 rounded border border-white/10 text-muted-foreground italic">
                    Invalid JSON
                  </div>
                )}
              </div>
            )}
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
          </div>
      </div>
    </div>
  )
})
LogRow.displayName = 'LogRow'

/**
 * LogViewer component.
 * @param props - The component props.
 * @param props.logs - The list of log entries to display.
 * @param props.highlightRegex - The regex to use for highlighting text.
 * @param props.isPaused - Whether the log stream is paused.
 * @returns The rendered component.
 */
export function LogViewer({ logs, highlightRegex, isPaused }: LogViewerProps) {
  return (
    <Suspense fallback={<div className="p-4 text-muted-foreground text-xs">Loading...</div>}>
      <Virtuoso
        style={{ height: '100%' }}
        data={logs}
        followOutput={isPaused ? false : 'auto'}
        className="p-4 scroll-smooth"
        itemContent={(_index: number, log: unknown) => (
          <LogRow key={(log as LogEntry).id} log={log as LogEntry} highlightRegex={highlightRegex} />
        )}
      />
    </Suspense>
  );
}
