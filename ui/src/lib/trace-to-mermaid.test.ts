import { describe, it, expect } from 'vitest';
import { traceToMermaid } from './trace-to-mermaid';
import { Trace } from '@/types/trace';

describe('traceToMermaid', () => {
  it('should generate a simple sequence for a single tool call', () => {
    const trace: Trace = {
      id: 't1',
      timestamp: new Date().toISOString(),
      totalDuration: 100,
      status: 'success',
      trigger: 'user',
      rootSpan: {
        id: 's1',
        name: 'test_tool',
        type: 'tool',
        startTime: 1000,
        endTime: 1100,
        status: 'success',
        serviceName: 'TestService'
      }
    };

    const diagram = traceToMermaid(trace);
    expect(diagram).toContain('participant User as User');
    expect(diagram).toContain('participant Core as MCP Any');
    expect(diagram).toContain('participant TestService as TestService');
    // Note: implementation maps root span directly to service if serviceName is present
    expect(diagram).toContain('User->>TestService: test_tool');
    expect(diagram).toContain('TestService-->>User: 100ms');
    // Let's check logic: processSpan(root, 'User').
    // root is mapped to 'TestService' (if serviceName present)?
    // Wait, my logic was: if serviceName present -> use that ID.
    // So rootSpan (serviceName=TestService) becomes TestService.
    // Parent=User. So User->>TestService.
    // Ah, logic:
    // if (span.serviceName) currentId = getSafeId(span.serviceName);
    // So root span IS the service execution.
    // User -> TestService.
    // Core is skipped if root has serviceName?
    // Let's check expectations against code.

    // Code:
    // processSpan(root, 'User')
    // root has serviceName='TestService'. currentId='TestService'.
    // User->>TestService.

    // If I want User->Core->Service, the root span should be Core, with a child that is Service.
    // But commonly traces start with the request.
    // If the trace object represents the whole flow, usually the top level is the Core handling.
    // If rootSpan has serviceName, it implies the Core delegated it? Or the span records the service interaction?
    // In many tracing systems, the root span is the incoming request (Core).
    // If my test data sets serviceName on root, it implies the root IS the service span.

    // Let's adjust expectation to match logic:
    // If root has serviceName, it is User->>TestService.
    // If root has NO serviceName (type=core), it is User->>Core.
  });

  it('should handle User -> Core -> Service flow', () => {
      // Correct trace structure for User -> Core -> Service
      const trace: Trace = {
      id: 't1',
      timestamp: new Date().toISOString(),
      totalDuration: 100,
      status: 'success',
      trigger: 'user',
      rootSpan: {
        id: 'root',
        name: 'gateway_request',
        type: 'core', // Executed by Core
        startTime: 1000,
        endTime: 1200,
        status: 'success',
        children: [
            {
                id: 'child',
                name: 'db_query',
                type: 'tool',
                serviceName: 'Postgres',
                startTime: 1050,
                endTime: 1150,
                status: 'success'
            }
        ]
      }
    };

    const diagram = traceToMermaid(trace);
    expect(diagram).toContain('participant User as User');
    expect(diagram).toContain('participant Core as MCP Any');
    expect(diagram).toContain('participant Postgres as Postgres');

    // User -> Core
    expect(diagram).toContain('User->>Core: gateway_request');

    // Core -> Postgres
    expect(diagram).toContain('Core->>Postgres: db_query');
    expect(diagram).toContain('Postgres-->>Core: 100ms');

    // Core -> User
    expect(diagram).toContain('Core-->>User: 200ms');
  });

  it('should handle nested calls', () => {
     const trace: Trace = {
      id: 't2',
      timestamp: new Date().toISOString(),
      totalDuration: 200,
      status: 'success',
      trigger: 'user',
      rootSpan: {
        id: 's1',
        name: 'root_op',
        type: 'core',
        startTime: 1000,
        endTime: 1200,
        status: 'success',
        children: [
            {
                id: 's2',
                name: 'sub_tool',
                type: 'tool',
                startTime: 1050,
                endTime: 1150,
                status: 'success',
                serviceName: 'SubService'
            }
        ]
      }
    };

    const diagram = traceToMermaid(trace);
    expect(diagram).toContain('participant SubService as SubService');
    expect(diagram).toContain('User->>Core: root_op');
    expect(diagram).toContain('Core->>SubService: sub_tool');
    expect(diagram).toContain('SubService-->>Core: 100ms');
  });
});
