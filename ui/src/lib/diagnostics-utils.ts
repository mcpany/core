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
