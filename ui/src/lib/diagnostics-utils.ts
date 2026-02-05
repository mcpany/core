/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Represents the result of a diagnostic analysis on a connection error.
 */
export interface DiagnosticResult {
  category: "network" | "auth" | "configuration" | "protocol" | "unknown";
  title: string;
  description: string;
  suggestion: string;
  severity: "critical" | "warning" | "info";
}

/**
 * Analyzes a raw connection error string and categorizes it into a user-friendly diagnostic result.
 *
 * @param error - The raw error string received from the backend or network.
 * @returns A structured DiagnosticResult containing the category, severity, and suggested remediation.
 */
export function analyzeConnectionError(error: string): DiagnosticResult {
  const err = error.toLowerCase();

  // Network: DNS / Host not found (Check this before generic "dial tcp")
  if (err.includes("no such host") || err.includes("name resolution failed")) {
    return {
      category: "configuration",
      title: "Host Not Found",
      description: "The hostname could not be resolved.",
      suggestion: "Check for typos in the hostname. Ensure the DNS is configured correctly within the container network.",
      severity: "critical",
    };
  }

  // Network: Connection Refused / Fetch Failed
  if (err.includes("connection refused") || err.includes("fetch failed") || (err.includes("dial tcp") && !err.includes("lookup"))) {
    return {
      category: "network",
      title: "Connection Refused",
      description: "The server is unreachable at the specified address/port.",
      suggestion: "1. Check if the upstream service is running.\n2. Verify the host and port are correct.\n3. Docker Users: 'localhost' refers to the container itself. Use 'host.docker.internal' to reach your host machine.",
      severity: "critical",
    };
  }

  // Network: Timeout
  if (err.includes("timeout") || err.includes("deadline exceeded")) {
    return {
      category: "network",
      title: "Connection Timeout",
      description: "The server took too long to respond.",
      suggestion: "Check firewall settings or network latency. The service might be hung or overloaded.",
      severity: "warning",
    };
  }

  // HTTP 404
  if (err.includes("404") || err.includes("not found")) {
    return {
      category: "configuration",
      title: "Not Found (404)",
      description: "The requested path or resource does not exist on the upstream server.",
      suggestion: "Check the URL path in your configuration. The base URL might be correct, but the endpoint might be wrong.",
      severity: "warning",
    };
  }

  // HTTP 500
  if (err.includes("500") || err.includes("internal server error")) {
    return {
      category: "protocol",
      title: "Internal Server Error (500)",
      description: "The upstream server encountered an error while processing the request.",
      suggestion: "Check the logs of the upstream service. It might be crashing or misconfigured.",
      severity: "critical",
    };
  }

  // HTTP 502/503
  if (err.includes("502") || err.includes("bad gateway") || err.includes("503") || err.includes("service unavailable")) {
    return {
      category: "network",
      title: "Service Unavailable (502/503)",
      description: "The upstream server is down or unable to handle the request.",
      suggestion: "The service might be restarting or overloaded. Check if the upstream process is running healthy.",
      severity: "critical",
    };
  }

  // Auth: Unauthorized
  if (err.includes("401") || err.includes("unauthorized") || err.includes("invalid token")) {
    return {
      category: "auth",
      title: "Authentication Failed",
      description: "The server rejected the credentials.",
      suggestion: "Verify your API key or Bearer token. Check if the token has expired.",
      severity: "critical",
    };
  }

  // Auth: Forbidden
  if (err.includes("403") || err.includes("forbidden")) {
    return {
      category: "auth",
      title: "Access Denied",
      description: "You are authenticated, but lack permissions.",
      suggestion: "Check if the token has the necessary scopes or permissions for this resource.",
      severity: "critical",
    };
  }

  // Mixed Content / WS mismatch
  if (err.includes("mixed content") || (err.includes("https") && err.includes("ws://"))) {
    return {
      category: "configuration",
      title: "Protocol Mismatch (Mixed Content)",
      description: "Attempting to connect to an insecure WebSocket (ws://) from a secure page (https://).",
      suggestion: "Browsers block this for security. Use 'wss://' (Secure WebSocket) or serve your dashboard over HTTP (not recommended).",
      severity: "critical",
    };
  }

  // Protocol: Handshake / TLS
  if (err.includes("handshake") || err.includes("certificate") || err.includes("tls")) {
    return {
      category: "protocol",
      title: "SSL/TLS Error",
      description: "Secure connection failed.",
      suggestion: "If using self-signed certificates, ensure the CA is trusted. Check if you are trying to connect via HTTPS to an HTTP port (or vice versa).",
      severity: "critical",
    };
  }

  // Protocol: Unexpected EOF
  if (err.includes("eof") || err.includes("closed connection")) {
    return {
      category: "protocol",
      title: "Connection Closed",
      description: "The upstream server closed the connection unexpectedly.",
      suggestion: "The server might have crashed or rejected the connection immediately. Check server logs.",
      severity: "warning",
    };
  }

  // Configuration: Schema Validation (Zod)
  if (err.includes("zod") || err.includes("validation failed") || err.includes("schema mismatch") || err.includes("invalid input")) {
    return {
      category: "configuration",
      title: "Schema Validation Error",
      description: "The upstream server returned data that does not match the expected schema.",
      suggestion: "1. Check if the upstream service API has changed (version mismatch).\n2. If using an adapter (e.g. Linear, Stripe), ensure the configuration matches the required inputs.\n3. Check for required environment variables that might be missing in the upstream service.",
      severity: "warning",
    };
  }

  // Configuration: Filesystem Permissions
  if (err.includes("access denied") || err.includes("eacces") || err.includes("permission denied")) {
    return {
      category: "configuration",
      title: "Permission Denied",
      description: "The server does not have permission to access a requested file or resource.",
      suggestion: "1. Check file system permissions.\n2. Ensure the 'allowed_paths' configuration covers the target directory.\n3. If on Windows, check for administrative privileges or path formatting.",
      severity: "critical",
    };
  }

  // Configuration: Missing Module
  if (err.includes("module_not_found") || err.includes("cannot find module")) {
    return {
      category: "configuration",
      title: "Missing Dependency",
      description: "The upstream service failed to load a required module.",
      suggestion: "1. Re-install dependencies (npm install).\n2. Check if the service requires a specific Node.js version.\n3. Verify the Docker image is built correctly.",
      severity: "critical",
    };
  }

  // Default / Unknown
  return {
    category: "unknown",
    title: "Unknown Error",
    description: error,
    suggestion: "Check the raw error message and server logs for more details.",
    severity: "warning",
  };
}
