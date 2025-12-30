
import React from 'react';
import { render } from '@testing-library/react';
import { ChartStyle, ChartConfig } from '../../src/components/ui/chart';
import { describe, it, expect } from 'vitest';

describe('ChartStyle Security', () => {
    it('should sanitize CSS values', () => {
        const config: ChartConfig = {
            test: {
                color: 'red; body { display: none; }',
            },
        };

        const { container } = render(<ChartStyle id="test-chart" config={config} />);
        const styleTag = container.querySelector('style');
        expect(styleTag).not.toBeNull();
        // The value should be sanitized
        expect(styleTag?.innerHTML).toContain('--color-test: red body  display: none ;');
        expect(styleTag?.innerHTML).not.toContain('red; body { display: none; }');
    });

    it('should sanitize CSS keys (VULNERABILITY REPRODUCTION)', () => {
        const maliciousKey = 'test: red; } body { display: none; } .bar';
        const config: ChartConfig = {
            [maliciousKey]: {
                color: 'blue',
            },
        };

        const { container } = render(<ChartStyle id="test-chart" config={config} />);
        const styleTag = container.querySelector('style');

        // After fix, the key should be sanitized
        expect(styleTag?.innerHTML).toContain('--color-test: red  body  display: none  .bar: blue;');

        // It should NOT contain the malicious payload that breaks out of the rule
        expect(styleTag?.innerHTML).not.toContain('} body { display: none; }');
    });
});
