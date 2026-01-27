/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { parseDuration, formatDuration } from "./duration-utils";
import Long from "long";

describe("duration-utils", () => {
  describe("parseDuration", () => {
    it("should parse seconds", () => {
      const d = parseDuration("10s");
      expect(d).toBeDefined();
      expect(Long.fromValue(d!.seconds).toNumber()).toBe(10);
      expect(d!.nanos).toBe(0);
    });

    it("should parse milliseconds", () => {
      const d = parseDuration("500ms");
      expect(d).toBeDefined();
      expect(Long.fromValue(d!.seconds).toNumber()).toBe(0);
      expect(d!.nanos).toBe(500000000);
    });

    it("should parse minutes", () => {
      const d = parseDuration("1m");
      expect(d).toBeDefined();
      expect(Long.fromValue(d!.seconds).toNumber()).toBe(60);
    });

    it("should parse fractional seconds", () => {
        const d = parseDuration("1.5s");
        expect(d).toBeDefined();
        expect(Long.fromValue(d!.seconds).toNumber()).toBe(1);
        expect(d!.nanos).toBe(500000000);
    });

    it("should return undefined for invalid input", () => {
      expect(parseDuration("invalid")).toBeUndefined();
      expect(parseDuration("10x")).toBeUndefined();
    });
  });

  describe("formatDuration", () => {
    it("should format seconds", () => {
      expect(formatDuration({ seconds: Long.fromNumber(10), nanos: 0 })).toBe("10s");
    });

    it("should format milliseconds", () => {
      expect(formatDuration({ seconds: Long.fromNumber(0), nanos: 500000000 })).toBe("500ms");
    });

    it("should format minutes", () => {
        expect(formatDuration({ seconds: Long.fromNumber(60), nanos: 0 })).toBe("1m");
    });

    it("should handle plain numbers for seconds", () => {
        expect(formatDuration({ seconds: 10, nanos: 0 })).toBe("10s");
    });
  });
});
