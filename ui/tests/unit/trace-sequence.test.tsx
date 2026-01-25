import { describe, it, expect } from 'vitest';
import { generateMermaidSequence } from '../../src/components/traces/trace-sequence';
import { Trace, Span } from '../../src/app/api/traces/route';

describe('generateMermaidSequence', () => {
    it('should generate a simple sequence for a single span', () => {
        const trace: Trace = {
            id: '1',
            rootSpan: {
                id: 'span1',
                name: 'test_tool',
                type: 'tool',
                startTime: 100,
                endTime: 200,
                status: 'success'
            },
            timestamp: '2023-01-01',
            totalDuration: 100,
            status: 'success',
            trigger: 'user'
        };

        const diagram = generateMermaidSequence(trace);
        expect(diagram).toContain('User->>test_tool: test_tool');
        expect(diagram).toContain('test_tool-->>User: OK');
    });

    it('should handle nested calls', () => {
        const trace: Trace = {
            id: '2',
            rootSpan: {
                id: 'span1',
                name: 'parent_tool',
                type: 'tool',
                startTime: 100,
                endTime: 500,
                status: 'success',
                children: [
                    {
                        id: 'span2',
                        name: 'child_service',
                        type: 'service',
                        serviceName: 'Database',
                        startTime: 200,
                        endTime: 300,
                        status: 'success'
                    }
                ]
            },
            timestamp: '2023-01-01',
            totalDuration: 400,
            status: 'success',
            trigger: 'user'
        };

        const diagram = generateMermaidSequence(trace);
        // Check for presence of key interactions
        expect(diagram).toContain('User->>parent_tool: parent_tool');
        expect(diagram).toContain('parent_tool->>Database: child_service');
        expect(diagram).toContain('Database-->>parent_tool: OK');
        expect(diagram).toContain('parent_tool-->>User: OK');
    });

    it('should sanitize names', () => {
        const trace: Trace = {
            id: '3',
            rootSpan: {
                id: 'span1',
                name: 'tool with spaces',
                type: 'tool',
                startTime: 100,
                endTime: 200,
                status: 'error',
                errorMessage: 'Failed'
            },
            timestamp: '2023-01-01',
            totalDuration: 100,
            status: 'error',
            trigger: 'user'
        };

        const diagram = generateMermaidSequence(trace);
        // "tool with spaces" sanitized becomes "tool_with_spaces"
        expect(diagram).toContain('participant tool_with_spaces as tool with spaces');
        expect(diagram).toContain('User->>tool_with_spaces: tool with spaces');
        expect(diagram).toContain('tool_with_spaces--xUser: ERR: Failed');
    });
});
