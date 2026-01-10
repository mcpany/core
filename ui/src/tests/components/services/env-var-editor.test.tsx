import { render, screen, fireEvent } from '@testing-library/react';
import { EnvVarEditor } from '@/components/services/env-var-editor';
import { describe, it, expect, vi } from 'vitest';

describe('EnvVarEditor', () => {
    it('renders correctly with no initial env vars', () => {
        const onChange = vi.fn();
        render(<EnvVarEditor onChange={onChange} />);
        expect(screen.getByText('Environment Variables')).toBeInTheDocument();
        expect(screen.getByText('No environment variables set.')).toBeInTheDocument();
    });

    it('renders initial env vars', () => {
        const initialEnv = {
            'TEST_VAR': { plainText: 'test_value' }
        };
        const onChange = vi.fn();
        render(<EnvVarEditor initialEnv={initialEnv} onChange={onChange} />);

        expect(screen.getByDisplayValue('TEST_VAR')).toBeInTheDocument();
        expect(screen.getByDisplayValue('test_value')).toBeInTheDocument();
    });

    it('renders and preserves secretId vars', () => {
        const initialEnv = {
            'SECRET_VAR': { secretId: 'my-secret-id' }
        };
        const onChange = vi.fn();
        render(<EnvVarEditor initialEnv={initialEnv} onChange={onChange} />);

        expect(screen.getByDisplayValue('SECRET_VAR')).toBeInTheDocument();
        // Should show as disabled input with secret ID
        expect(screen.getByDisplayValue('Secret: my-secret-id')).toBeInTheDocument();
        expect(screen.getByDisplayValue('Secret: my-secret-id')).toBeDisabled();
    });

    it('adds a new variable', () => {
        const onChange = vi.fn();
        render(<EnvVarEditor onChange={onChange} />);

        const addButton = screen.getByText('Add Variable');
        fireEvent.click(addButton);

        const inputs = screen.getAllByRole('textbox');
        expect(inputs.length).toBeGreaterThan(0);
    });

    it('updates a variable and calls onChange', () => {
        const onChange = vi.fn();
        render(<EnvVarEditor onChange={onChange} />);

        fireEvent.click(screen.getByText('Add Variable'));

        const keyInput = screen.getByPlaceholderText('KEY');
        fireEvent.change(keyInput, { target: { value: 'NEW_KEY' } });

        expect(onChange).toHaveBeenCalledWith({
            'NEW_KEY': { plainText: '' }
        });

        const valueInput = screen.getByPlaceholderText('VALUE');
        fireEvent.change(valueInput, { target: { value: 'new_val' } });

         expect(onChange).toHaveBeenCalledWith({
            'NEW_KEY': { plainText: 'new_val' }
        });
    });

    it('removes a variable', () => {
        const initialEnv = {
            'VAR1': { plainText: 'val1' }
        };
        const onChange = vi.fn();
        render(<EnvVarEditor initialEnv={initialEnv} onChange={onChange} />);

        const allButtons = screen.getAllByRole('button');
        const deleteButton = allButtons[allButtons.length - 1];

        fireEvent.click(deleteButton);

        expect(onChange).toHaveBeenCalledWith({});
        expect(screen.getByText('No environment variables set.')).toBeInTheDocument();
    });
});
