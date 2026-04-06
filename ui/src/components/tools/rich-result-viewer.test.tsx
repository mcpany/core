/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { RichResultViewer } from './rich-result-viewer';

describe('RichResultViewer', () => {
    it('delegates to SmartResultRenderer and displays JSON mode initially', () => {
        const result = { test: 'value' };
        render(<RichResultViewer result={result} />);

        // Ensure "tab" is returned for UI e2e compat
        const jsonButton = screen.getByRole('tab', { name: /JSON/i });
        expect(jsonButton).toBeDefined();
    });
});
