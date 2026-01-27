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

  // 1. Schema Validation Errors
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

  // 5. Connection Errors
  if (
    lowerMsg.includes('connection refused') ||
    lowerMsg.includes('failed to connect') ||
    lowerMsg.includes('econnrefused')
  ) {
    diagnostics.push({
      type: 'error',
      title: 'Connection Failed',
      message: 'Could not connect to the upstream service.',
      suggestion: 'Ensure the upstream service is running and accessible from the MCP Any server container.'
    });
  }

  // Fallback if we have an error but matched nothing
  if (diagnostics.length === 0 && errorMessage) {
      diagnostics.push({
          type: 'error',
          title: 'Unknown Error',
          message: errorMessage,
          suggestion: 'Review the logs for more context.'
      });
  }

  return diagnostics;
}
