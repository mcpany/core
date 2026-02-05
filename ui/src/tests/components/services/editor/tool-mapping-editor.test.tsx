/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { ToolMappingEditor } from '@/components/services/editor/tool-mapping-editor';
import { describe, it, expect, vi } from 'vitest';
import { UpstreamServiceConfig } from '@/lib/client';

const mockService: UpstreamServiceConfig = {
    id: 'test-service',
    name: 'test-service',
    version: '1.0.0',
    disable: false,
    priority: 0,
    loadBalancingStrategy: 0,
    sanitizedName: 'test-service',
    httpService: {
        address: 'http://localhost',
        tools: [],
        calls: {},
        resources: [],
        prompts: []
    },
    configurationSchema: "",
    tags: [],
    callPolicies: [],
    preCallHooks: [],
    postCallHooks: [],
    prompts: [],
    autoDiscoverTool: false,
    configError: "",
    readOnly: false,
    toolExportPolicy: undefined,
    promptExportPolicy: undefined,
    resourceExportPolicy: undefined
};

describe('ToolMappingEditor', () => {
    it('renders correctly', () => {
        const onChange = vi.fn();
        render(<ToolMappingEditor service={mockService} onChange={onChange} />);
        expect(screen.getByText('Tool Mappings')).toBeInTheDocument();
        expect(screen.getByText('No tools defined.')).toBeInTheDocument();
    });

    it('opens add tool dialog', () => {
        const onChange = vi.fn();
        render(<ToolMappingEditor service={mockService} onChange={onChange} />);

        fireEvent.click(screen.getByText('Add Tool'));
        expect(screen.getByText('Tool Interface')).toBeInTheDocument();
        expect(screen.getByText('Upstream Call')).toBeInTheDocument();
    });

    it('saves a new tool', () => {
        const onChange = vi.fn();
        render(<ToolMappingEditor service={mockService} onChange={onChange} />);

        fireEvent.click(screen.getByText('Add Tool'));

        const nameInput = screen.getByLabelText('Name');
        fireEvent.change(nameInput, { target: { value: 'get_weather' } });

        const pathInput = screen.getByLabelText('Endpoint Path');
        fireEvent.change(pathInput, { target: { value: '/weather' } });

        const saveButton = screen.getByText('Save Tool');
        fireEvent.click(saveButton);

        expect(onChange).toHaveBeenCalled();
        const callArgs = onChange.mock.calls[0][0];
        expect(callArgs.httpService.tools).toHaveLength(1);
        expect(callArgs.httpService.tools[0].name).toBe('get_weather');

        const callId = callArgs.httpService.tools[0].callId;
        expect(callArgs.httpService.calls[callId]).toBeDefined();
        expect(callArgs.httpService.calls[callId].endpointPath).toBe('/weather');
    });
});
