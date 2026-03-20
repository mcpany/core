import { render, screen, fireEvent, within } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { StepParameters } from './step-parameters';
import * as wizardContext from '../wizard-context';

// Mock context correctly to prevent invariant violations and control returns
vi.mock('../wizard-context', () => ({
  useWizard: vi.fn(),
}));

describe('StepParameters Component', () => {
    const mockUpdateState = vi.fn();
    const mockUpdateConfig = vi.fn();

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('correctly adds and syncs an environment variable', () => {
        const mockContextValue = {
            state: {
                params: {
                   "INITIAL_VAR": "initial_value"
                },
                config: {
                    commandLineService: {
                        command: 'test-command',
                        env: { "INITIAL_VAR": { plainText: "initial_value"} }
                    }
                }
            },
            updateState: mockUpdateState,
            updateConfig: mockUpdateConfig,
        };

        vi.mocked(wizardContext.useWizard).mockReturnValue(mockContextValue as any);

        render(<StepParameters />);

        const addParamBtn = screen.getByText('Add Parameter');
        fireEvent.click(addParamBtn);

        expect(mockUpdateState).toHaveBeenCalledWith({
           params: { "INITIAL_VAR": "initial_value", "": "" }
        });
    });

    it('updates parameter key and syncs properly, ignoring empty keys', () => {
         const mockContextValue = {
            state: {
                params: {
                   "": ""
                },
                config: {
                    commandLineService: {
                        command: 'test-command',
                        env: {}
                    }
                }
            },
            updateState: mockUpdateState,
            updateConfig: mockUpdateConfig,
        };

        vi.mocked(wizardContext.useWizard).mockReturnValue(mockContextValue as any);

        render(<StepParameters />);

        // Get inputs (first is key, second is value)
        const inputs = screen.getAllByRole('textbox');
        const keyInput = inputs[0];

        // Type a key
        fireEvent.change(keyInput, { target: { value: 'NEW_KEY' }});

        // The parameter change expects updateState with changed params
        expect(mockUpdateState).toHaveBeenCalledWith({
           params: { "NEW_KEY": "" }
        });

        // The config update expects to set 'plainText'
        expect(mockUpdateConfig).toHaveBeenCalledWith({
            commandLineService: {
                command: 'test-command',
                env: {
                    "NEW_KEY": { plainText: "" }
                }
            }
        });
    });

    it('removes a parameter and syncs correctly', () => {
          const mockContextValue = {
            state: {
                params: {
                   "TO_REMOVE": "value"
                },
                config: {
                    commandLineService: {
                        command: 'test-command',
                        env: {
                            "TO_REMOVE": { plainText: "value"}
                        }
                    }
                }
            },
            updateState: mockUpdateState,
            updateConfig: mockUpdateConfig,
        };

        vi.mocked(wizardContext.useWizard).mockReturnValue(mockContextValue as any);

        render(<StepParameters />);

        // Get the table rows and find the button within the one with "TO_REMOVE"
        const rows = screen.getAllByRole('row');
        // The first row is the header, the second is our param
        const dataRow = rows[1];

        // Get the remove button (Trash2 icon inside) inside this row
        const removeBtn = within(dataRow).getByRole('button');
        fireEvent.click(removeBtn);

        expect(mockUpdateState).toHaveBeenCalledWith({
           params: {}
        });

        // The config update expects to empty env
        expect(mockUpdateConfig).toHaveBeenCalledWith({
            commandLineService: {
                command: 'test-command',
                env: {}
            }
        });
    });
});
