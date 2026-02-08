import { describe, it, expect } from 'vitest';
import { parseDuration, formatDuration, formatDurationForApi } from './duration-utils';
import Long from 'long';

describe('duration-utils', () => {
    describe('parseDuration', () => {
        it('parses seconds', () => {
            const d = parseDuration('30s');
            expect(d).toBeDefined();
            expect(d?.seconds.toNumber()).toBe(30);
            expect(d?.nanos).toBe(0);
        });

        it('parses milliseconds', () => {
            const d = parseDuration('500ms');
            expect(d).toBeDefined();
            expect(d?.seconds.toNumber()).toBe(0);
            expect(d?.nanos).toBe(500000000);
        });

        it('parses fractional seconds', () => {
            const d = parseDuration('1.5s');
            expect(d).toBeDefined();
            expect(d?.seconds.toNumber()).toBe(1);
            expect(d?.nanos).toBe(500000000);
        });

        it('parses minutes', () => {
            const d = parseDuration('2m');
            expect(d).toBeDefined();
            expect(d?.seconds.toNumber()).toBe(120);
            expect(d?.nanos).toBe(0);
        });

        it('returns undefined for invalid format', () => {
            expect(parseDuration('invalid')).toBeUndefined();
            expect(parseDuration('10')).toBeUndefined(); // Missing unit
        });
    });

    describe('formatDuration', () => {
        it('formats seconds', () => {
            const d = { seconds: Long.fromNumber(30), nanos: 0 };
            expect(formatDuration(d)).toBe('30s');
        });

        it('formats milliseconds', () => {
            const d = { seconds: Long.fromNumber(0), nanos: 500000000 };
            expect(formatDuration(d)).toBe('500ms');
        });

        it('formats mixed', () => {
            const d = { seconds: Long.fromNumber(1), nanos: 500000000 };
            expect(formatDuration(d)).toBe('1.5s');
        });

        it('formats minutes if exact', () => {
             const d = { seconds: Long.fromNumber(120), nanos: 0 };
             expect(formatDuration(d)).toBe('2m');
        });

        it('formats minutes as seconds if not exact', () => {
             const d = { seconds: Long.fromNumber(121), nanos: 0 };
             expect(formatDuration(d)).toBe('121s');
        });
    });

    describe('formatDurationForApi', () => {
        it('always formats as seconds', () => {
            expect(formatDurationForApi({ seconds: Long.fromNumber(30), nanos: 0 })).toBe('30s');
            expect(formatDurationForApi({ seconds: Long.fromNumber(0), nanos: 500000000 })).toBe('0.5s');
            expect(formatDurationForApi({ seconds: Long.fromNumber(120), nanos: 0 })).toBe('120s');
            expect(formatDurationForApi({ seconds: Long.fromNumber(0), nanos: 1000000 })).toBe('0.001s');
        });

        it('handles small durations without scientific notation', () => {
             // 1 microsecond = 0.000001s
             expect(formatDurationForApi({ seconds: Long.fromNumber(0), nanos: 1000 })).toBe('0.000001s');
             // 1 nanosecond = 0.000000001s
             expect(formatDurationForApi({ seconds: Long.fromNumber(0), nanos: 1 })).toBe('0.000000001s');
        });

        it('handles undefined/null', () => {
            expect(formatDurationForApi(undefined)).toBeUndefined();
            expect(formatDurationForApi(null)).toBeUndefined();
        });
    });
});
