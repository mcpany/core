/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { StepParameters } from './step-parameters';
import { vi, describe, it, expect } from 'vitest';
import { useWizard } from '../wizard-context';

// Mock useWizard
vi.mock('../wizard-context', () => ({
  useWizard: vi.fn()
}));

describe('StepParameters', () => {
  it('renders schema form when schema is present', () => {
    (useWizard as any).mockReturnValue({
        state: {
            config: {
                configurationSchema: JSON.stringify({
                    properties: {
                        TEST_KEY: { type: 'string', title: 'Test Key' }
                    }
                }),
                commandLineService: { env: {} }
            },
            params: {}
        },
        updateState: vi.fn(),
        updateConfig: vi.fn()
    });

    render(<StepParameters />);
    expect(screen.getByLabelText('Test Key')).toBeInTheDocument();
  });

  it('renders manual editor when schema is missing', () => {
    (useWizard as any).mockReturnValue({
        state: {
            config: {
                commandLineService: { env: {} }
            },
            params: {}
        },
        updateState: vi.fn(),
        updateConfig: vi.fn()
    });

    render(<StepParameters />);
    expect(screen.getByText('Add Parameter')).toBeInTheDocument();
  });
});
