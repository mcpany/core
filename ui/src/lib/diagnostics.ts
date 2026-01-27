/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Trace } from "@/types/trace";

/**
 * Diagnostic represents a finding from the trace analysis, indicating an error or warning.
 */
export interface Diagnostic {
  type: 'error' | 'warning' | 'info';
  title: string;
  message: string;
  suggestion?: string;
}

/**
 * Analyzes a trace for common errors and returns a list of diagnostics.
 * @param trace - The trace to analyze.
 * @returns An array of diagnostics.
 */
export function analyzeTrace(trace: Trace): Diagnostic[] {
  const diagnostics: Diagnostic[] = [];
  const rootSpan = trace.rootSpan;

  if (trace.status !== 'error' && rootSpan.status !== 'error') {
    return diagnostics;
  }

  const errorMessage = rootSpan.errorMessage ||
                       (typeof rootSpan.output?.error === 'string' ? rootSpan.output.error : null) ||
                       (typeof rootSpan.output?.message === 'string' ? rootSpan.output.message : null) ||
                       '';

  const lowerMsg = errorMessage.toLowerCase();

  // Check for Recursion Depth
  const getMaxDepth = (span: any, currentDepth = 1): number => {
      if (!span.children || span.children.length === 0) return currentDepth;
      return Math.max(...span.children.map((child: any) => getMaxDepth(child, currentDepth + 1)));
  };
  const depth = getMaxDepth(rootSpan);

  if (depth > 10) {
      diagnostics.push({
          type: 'warning',
          title: 'High Recursion Depth',
          message: `The trace depth is ${depth}, which might indicate an infinite loop or inefficient chain.`,
          suggestion: 'Check if the agent is stuck in a loop calling the same tools repeatedly.'
      });
  }

  // 1. Authentication & Authorization Errors
  if (
      lowerMsg.includes('401') ||
      lowerMsg.includes('403') ||
      lowerMsg.includes('unauthorized') ||
      lowerMsg.includes('unauthenticated') ||
      lowerMsg.includes('invalid api key') ||
      lowerMsg.includes('invalid token')
  ) {
      diagnostics.push({
          type: 'error',
          title: 'Authentication Failed',
          message: 'The server rejected the credentials provided.',
          suggestion: 'Go to Settings > Secrets and ensure the API Key for this service is correct and up to date.'
      });
  }

  // 2. Missing Tool Errors
  if (
      lowerMsg.includes('tool not found') ||
      lowerMsg.includes('unknown tool') ||
      lowerMsg.includes('method not found')
  ) {
      diagnostics.push({
          type: 'error',
          title: 'Tool Not Found',
          message: 'The requested tool does not exist or is not enabled.',
          suggestion: 'Check the "Services" page to ensure the service providing this tool is enabled. Verify the tool name in the "Tools" list.'
      });
  }

  // 3. Rate Limit Errors
  if (
      lowerMsg.includes('429') ||
      lowerMsg.includes('rate limit') ||
      lowerMsg.includes('quota exceeded') ||
      lowerMsg.includes('too many requests')
  ) {
      diagnostics.push({
          type: 'error',
          title: 'Rate Limit Exceeded',
          message: 'The upstream service is rejecting requests due to high volume.',
          suggestion: 'Implement caching or exponential backoff. Consider upgrading the upstream service plan if applicable.'
      });
  }

  // 4. Schema Validation Errors
  if (
    lowerMsg.includes('schema validation') ||
    lowerMsg.includes('validation error') ||
    lowerMsg.includes('invalid input') ||
    lowerMsg.includes('zoderror')
  ) {
    diagnostics.push({
      type: 'error',
      title: 'Schema Validation Error',
      message: 'The tool arguments did not match the expected schema.',
      suggestion: 'Check the "Payload" tab to see the input arguments. Compare them against the tool definition in the "Tools" page.'
    });
  }

  // 2. Permission Errors
  if (
    lowerMsg.includes('eperm') ||
    lowerMsg.includes('eacces') ||
    lowerMsg.includes('permission denied') ||
    lowerMsg.includes('access denied')
  ) {
    diagnostics.push({
      type: 'error',
      title: 'Permission Denied',
      message: 'The server does not have permission to perform this action.',
      suggestion: 'If using the filesystem server, ensure it has access to the target directory. You may need to run the server with different permissions or update the configuration.'
    });
  }

  // 3. JSON Parsing Errors
  if (
    lowerMsg.includes('json parse error') ||
    lowerMsg.includes('syntaxerror') ||
    lowerMsg.includes('unexpected token')
  ) {
    diagnostics.push({
      type: 'error',
      title: 'JSON Parsing Error',
      message: 'The server received or produced invalid JSON.',
      suggestion: 'Check the upstream service response. It might be returning HTML or non-JSON data.'
    });
  }

  // 4. Timeout Errors
  if (
    lowerMsg.includes('timeout') ||
    lowerMsg.includes('deadline exceeded')
  ) {
    diagnostics.push({
      type: 'error',
      title: 'Operation Timed Out',
      message: 'The operation took too long to complete.',
      suggestion: 'The upstream service might be slow or unresponsive. Try increasing the timeout in the server configuration.'
    });
  }

  // 8. Connection Errors
  if (
    lowerMsg.includes('connection refused') ||
    lowerMsg.includes('failed to connect') ||
    lowerMsg.includes('econnrefused') ||
    lowerMsg.includes('network error')
  ) {
    diagnostics.push({
      type: 'error',
      title: 'Connection Failed',
      message: 'Could not connect to the upstream service.',
      suggestion: 'Ensure the upstream service is running and accessible from the MCP Any server container. Check firewall rules and network policies.'
    });
  }

  // Fallback if we have an error but matched nothing
  const hasErrorDiagnostic = diagnostics.some(d => d.type === 'error');
  if (!hasErrorDiagnostic && errorMessage) {
      diagnostics.push({
          type: 'error',
          title: 'Unknown Error',
          message: errorMessage,
          suggestion: 'Review the logs for more context.'
      });
  }

  return diagnostics;
}
