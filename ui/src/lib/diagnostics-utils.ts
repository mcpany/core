/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

export interface DiagnosticResult {
  category: "network" | "auth" | "configuration" | "protocol" | "unknown";
  title: string;
  description: string;
  suggestion: string;
  severity: "critical" | "warning" | "info";
}

/**
 * analyzeConnectionError.
 *
 * @param error - The error.
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

  // Network: Connection Refused
  if (err.includes("connection refused") || (err.includes("dial tcp") && !err.includes("lookup"))) {
    return {
      category: "network",
      title: "Connection Refused",
      description: "The server is unreachable at the specified address/port.",
      suggestion: "1. Check if the upstream service is running.\n2. Verify the host and port are correct.\n3. If running in Docker, ensure you are using the correct network alias (e.g., 'host.docker.internal' instead of 'localhost').",
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

  // Default / Unknown
  return {
    category: "unknown",
    title: "Unknown Error",
    description: error,
    suggestion: "Check the raw error message and server logs for more details.",
    severity: "warning",
  };
}
