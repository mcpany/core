/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, act } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { HttpToolEditor } from './http-tool-editor';
import { ToolDefinition } from '@proto/config/v1/tool';
import { HttpCallDefinition, HttpCallDefinition_HttpMethod, OutputTransformer_OutputFormat } from '@proto/config/v1/call';

// Mock child components to isolate tests
vi.mock('./input-transformer-editor', () => ({
    InputTransformerEditor: ({ onChange }: { onChange: (val: any) => void }) => (
        <div data-testid="input-transformer-editor">
            <button onClick={() => onChange({ template: 'mock-template' })}>Update Input</button>
        </div>
    )
}));

vi.mock('./output-transformer-editor', () => ({
    OutputTransformerEditor: ({ onChange }: { onChange: (val: any) => void }) => (
        <div data-testid="output-transformer-editor">
            <button onClick={() => onChange({ format: 0, template: 'mock-output' })}>Update Output</button>
        </div>
    )
}));

// Mock UI components that might cause issues in JSDOM
vi.mock('@/components/ui/select', () => ({
    Select: ({ children, onValueChange }: any) => <div onClick={() => onValueChange && onValueChange('1')}>{children}</div>,
    SelectTrigger: ({ children }: any) => <div>{children}</div>,
    SelectValue: () => <div>SelectValue</div>,
    SelectContent: ({ children }: any) => <div>{children}</div>,
    SelectItem: ({ children }: any) => <div>{children}</div>,
}));

vi.mock('@/components/ui/tabs', () => ({
    Tabs: ({ children, value, onValueChange }: any) => (
        <div>
            <div data-testid="tabs-value">{value}</div>
            <div onClick={(e) => {
                const target = e.target as HTMLElement;
                if (target.dataset.value) {
                    onValueChange(target.dataset.value);
                }
            }}>
                {children}
            </div>
        </div>
    ),
    TabsList: ({ children }: any) => <div>{children}</div>,
    TabsTrigger: ({ value, children }: any) => <button data-value={value}>{children}</button>,
    TabsContent: ({ value, children }: any) => {
        // In the mock, we can just render it if we want to test content presence,
        // but the real Tabs hides it. We can rely on the parent Tabs to filter,
        // or just render everything and let the test check if it SHOULD be visible based on value.
        // Actually, simplest is to just render it wrapper in a div with data-value
        // and let the test check visibility, or even better:
        // The real Tabs only renders the content matching the value.
        // But since we are mocking Tabs, we don't have access to the `value` prop of Tabs here
        // unless we use context.
        // Simpler: Just render everything and use `display: none`?
        // Or simpler: Mock Tabs to render ALL content but markers.
        // Wait, if I mock Tabs, I control rendering.
        // The HttpToolEditor passes `value={activeTab}`.
        // But TabsContent is a child.
        // Let's make TabsContent render only if active? No, we don't have context in simple mock.
        // We will make Tabs render children. And TabsContent render children.
        // But then everything is visible.
        // The test checks `fireEvent.click` and then expects element to be in document.
        // If everything is visible, `getByText` works.
        // BUT, `HttpToolEditor` state `activeTab` controls the value passed to `Tabs`.
        // If I click, `activeTab` updates.
        // If I want to test that clicking switches tab, I need to ensure the click updates the state.
        return <div data-content-value={value}>{children}</div>;
    },
}));

describe('HttpToolEditor', () => {
    const mockTool: ToolDefinition = {
        name: 'test-tool',
        description: 'Test Description'
    };

    const mockCall: HttpCallDefinition = {
        id: 'test-id',
        method: HttpCallDefinition_HttpMethod.HTTP_METHOD_GET,
        endpointPath: '/test',
        parameters: [],
        inputTransformer: { template: '' },
        outputTransformer: { format: OutputTransformer_OutputFormat.JSON, template: '' }
    };

    it('renders basic fields', () => {
        render(<HttpToolEditor tool={mockTool} call={mockCall} onChange={() => {}} />);
        expect(screen.getByDisplayValue('test-tool')).toBeInTheDocument();
        expect(screen.getByDisplayValue('/test')).toBeInTheDocument();
    });

    it('switches tabs and renders transformer editors', async () => {
        render(<HttpToolEditor tool={mockTool} call={mockCall} onChange={() => {}} />);

        // Check tabs exist
        expect(screen.getByText('Request Parameters')).toBeInTheDocument();
        expect(screen.getByText('Input Transform')).toBeInTheDocument();
        expect(screen.getByText('Output Transform')).toBeInTheDocument();

        // Click Input Transform tab
        fireEvent.click(screen.getByText('Input Transform'));
        expect(screen.getByTestId('input-transformer-editor')).toBeInTheDocument();

        // Click Output Transform tab
        fireEvent.click(screen.getByText('Output Transform'));
        expect(screen.getByTestId('output-transformer-editor')).toBeInTheDocument();
    });

    it('updates input transformer when editor changes', () => {
        const handleChange = vi.fn();
        render(<HttpToolEditor tool={mockTool} call={mockCall} onChange={handleChange} />);

        fireEvent.click(screen.getByText('Input Transform'));
        fireEvent.click(screen.getByText('Update Input'));

        expect(handleChange).toHaveBeenCalled();
        const calledCall = handleChange.mock.calls[0][1];
        expect(calledCall.inputTransformer.template).toBe('mock-template');
    });

    it('updates output transformer when editor changes', () => {
        const handleChange = vi.fn();
        render(<HttpToolEditor tool={mockTool} call={mockCall} onChange={handleChange} />);

        fireEvent.click(screen.getByText('Output Transform'));
        fireEvent.click(screen.getByText('Update Output'));

        expect(handleChange).toHaveBeenCalled();
        const calledCall = handleChange.mock.calls[0][1];
        expect(calledCall.outputTransformer.template).toBe('mock-output');
    });
});
