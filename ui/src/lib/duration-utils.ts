/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import Long from "long";

interface Duration {
  seconds: Long | number | string;
  nanos: number;
}

/**
 * Parses a duration string (e.g., "1.5s", "100ms") into a Protobuf Duration object.
 * Supports: ns, us, ms, s, m, h.
 * @param input The duration string.
 * @returns The Duration object or undefined if invalid.
 */
export function parseDuration(input: string): Duration | undefined {
  if (!input) return undefined;

  const regex = /^([\d.]+)(ns|us|µs|ms|s|m|h)?$/;
  const match = input.trim().match(regex);

  if (!match) return undefined;

  let value = parseFloat(match[1]);
  const unit = match[2] || "s";

  if (isNaN(value)) return undefined;

  // Normalize to seconds
  switch (unit) {
    case "ns":
      value /= 1e9;
      break;
    case "us":
    case "µs":
      value /= 1e6;
      break;
    case "ms":
      value /= 1e3;
      break;
    case "s":
      break;
    case "m":
      value *= 60;
      break;
    case "h":
      value *= 3600;
      break;
  }

  const seconds = Math.floor(value);
  const nanos = Math.round((value - seconds) * 1e9);

  return {
    seconds: Long.fromNumber(seconds),
    nanos: nanos,
  };
}

/**
 * Formats a Protobuf Duration object into a human-readable string.
 * @param duration The Duration object.
 * @returns The formatted string (e.g., "1.5s") or empty string if undefined.
 */
export function formatDuration(duration: Duration | undefined): string {
  if (!duration) return "";

  const seconds = Long.isLong(duration.seconds)
    ? (duration.seconds as Long).toNumber()
    : Number(duration.seconds);
  const nanos = duration.nanos || 0;

  if (seconds === 0 && nanos === 0) return "0s";

  const totalSeconds = seconds + nanos / 1e9;

  // Smart formatting
  if (totalSeconds < 1e-6) return `${totalSeconds * 1e9}ns`;
  if (totalSeconds < 1e-3) return `${totalSeconds * 1e6}us`;
  if (totalSeconds < 1) return `${totalSeconds * 1e3}ms`;
  if (totalSeconds % 60 === 0 && totalSeconds >= 60) {
      if (totalSeconds % 3600 === 0) return `${totalSeconds / 3600}h`;
      return `${totalSeconds / 60}m`;
  }

  return `${totalSeconds}s`;
}
