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

        // SmartResultRenderer displays a JSON button when it renders raw JSON
        const jsonButtons = screen.getAllByRole('button', { name: /JSON/i });
        expect(jsonButtons.length).toBeGreaterThan(0);
    });
});
