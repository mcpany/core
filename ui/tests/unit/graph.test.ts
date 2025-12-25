
import { describe, it, expect } from 'vitest';
import { transformServicesToGraph } from '../../src/lib/graph-utils';
import { Service } from '../../src/types/service';

describe('transformServicesToGraph', () => {
  it('should create a central node and service nodes', () => {
    const services: Service[] = [
      {
        id: 'svc1',
        name: 'Test Service 1',
        version: 'v1',
        disable: false,
        service_config: { http_service: { address: 'localhost' } }
      },
      {
        id: 'svc2',
        name: 'Test Service 2',
        version: 'v2',
        disable: true,
        service_config: { grpc_service: { address: 'localhost' } }
      }
    ];

    const { nodes, edges } = transformServicesToGraph(services);

    // Should have 1 central node + 2 service nodes = 3 nodes
    expect(nodes).toHaveLength(3);

    // Check Central Node
    const centralNode = nodes.find(n => n.id === 'mcp-any');
    expect(centralNode).toBeDefined();
    expect(centralNode?.type).toBe('central');

    // Check Service Nodes
    const svc1Node = nodes.find(n => n.id === 'svc1');
    expect(svc1Node).toBeDefined();
    expect(svc1Node?.data.label).toBe('Test Service 1');
    expect(svc1Node?.data.status).toBe('active');
    expect(svc1Node?.data.type).toBe('HTTP');

    const svc2Node = nodes.find(n => n.id === 'svc2');
    expect(svc2Node).toBeDefined();
    expect(svc2Node?.data.status).toBe('disabled');
    expect(svc2Node?.data.type).toBe('gRPC');

    // Should have 2 edges connecting center to services
    expect(edges).toHaveLength(2);
    expect(edges.find(e => e.source === 'mcp-any' && e.target === 'svc1')).toBeDefined();
    expect(edges.find(e => e.source === 'mcp-any' && e.target === 'svc2')).toBeDefined();
  });
});
