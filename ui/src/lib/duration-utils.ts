import { Duration } from "@proto/google/protobuf/duration";
import Long from "long";

/**
 * Parses a duration string (e.g. "1.5s", "500ms") into a Protobuf Duration object.
 * @param input The duration string.
 * @returns The Duration object or undefined if invalid.
 */
export function parseDuration(input: string): Duration | undefined {
    if (!input) return undefined;

    // Simple regex for common duration formats
    const match = input.match(/^([\d.]+)(ns|us|ms|s|m|h)$/);
    if (!match) {
        // Fallback: if purely numeric, assume seconds? No, better to require unit or fail.
        // Actually, some APIs treat raw number as seconds. Let's support it if needed, but safe default is undefined.
        return undefined;
    }

    const val = parseFloat(match[1]);
    const unit = match[2];

    if (isNaN(val)) return undefined;

    let totalSeconds = 0;
    switch (unit) {
        case 'ns': totalSeconds = val / 1e9; break;
        case 'us': totalSeconds = val / 1e6; break;
        case 'ms': totalSeconds = val / 1000; break;
        case 's': totalSeconds = val; break;
        case 'm': totalSeconds = val * 60; break;
        case 'h': totalSeconds = val * 3600; break;
    }

    const seconds = Math.floor(totalSeconds);
    const nanos = Math.round((totalSeconds - seconds) * 1e9);

    return {
        seconds: Long.fromNumber(seconds),
        nanos: nanos
    };
}

/**
 * Formats a Protobuf Duration object into a human-readable string.
 * @param duration The Duration object.
 * @returns The formatted string (e.g. "1.5s", "500ms").
 */
export function formatDuration(duration: Duration | undefined | null): string {
    if (!duration) return "";

    // Handle Long or string or number type for seconds
    let seconds = 0;
    if (Long.isLong(duration.seconds)) {
        seconds = duration.seconds.toNumber();
    } else if (typeof duration.seconds === 'string') {
        seconds = parseInt(duration.seconds, 10);
    } else if (typeof duration.seconds === 'number') {
        seconds = duration.seconds;
    }

    const nanos = duration.nanos || 0;

    if (seconds === 0 && nanos === 0) return "0s";

    const totalSeconds = seconds + nanos / 1e9;

    // Smart formatting
    if (totalSeconds < 0.000001) return `${totalSeconds * 1e9}ns`;
    if (totalSeconds < 0.001) return `${totalSeconds * 1e6}us`;
    if (totalSeconds < 1) return `${totalSeconds * 1000}ms`;
    if (totalSeconds >= 60 && totalSeconds % 60 === 0) return `${totalSeconds / 60}m`;

    return `${totalSeconds}s`;
}

/**
 * Formats a Protobuf Duration object into a string compliant with Google Protobuf JSON mapping (always seconds).
 * @param duration The Duration object.
 * @returns The formatted string (e.g. "1.5s", "0.5s") or undefined if input is null/undefined.
 */
export function formatDurationForApi(duration: Duration | undefined | null): string | undefined {
    if (!duration) return undefined;

    // Handle Long or string or number type for seconds
    let seconds = 0;
    if (Long.isLong(duration.seconds)) {
        seconds = duration.seconds.toNumber();
    } else if (typeof duration.seconds === 'string') {
        seconds = parseInt(duration.seconds, 10);
    } else if (typeof duration.seconds === 'number') {
        seconds = duration.seconds;
    }

    const nanos = duration.nanos || 0;

    // Total seconds as float
    const totalSeconds = seconds + nanos / 1e9;

    // Format with up to 9 decimal places to support nanoseconds
    // Protobuf JSON mapping uses 0, 3, 6, or 9 fractional digits.
    // We can just print as number and append 's', but avoiding scientific notation is safer.
    // Using toFixed(9) and stripping zeros covers it.
    let s = totalSeconds.toFixed(9);
    // Remove trailing zeros and decimal point if integer
    s = s.replace(/\.?0+$/, "");
    return `${s}s`;
}
