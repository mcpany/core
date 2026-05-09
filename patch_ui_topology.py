import re

filename = "ui/src/components/dashboard/swarm-topology-widget.tsx"
with open(filename, 'r') as f:
    content = f.read()

# Replace the mock fetch block
old_fetch = """        const generateMockData = (): SwarmTopologyData => {
            const nodes: SwarmNode[] = [
                { id: 'n1', label: 'Primary Orchestrator', type: 'validator', status: 'locked', x: 50, y: 50 },
                { id: 'n2', label: 'Research Agent', type: 'agent', status: 'active', x: 20, y: 30 },
                { id: 'n3', label: 'Tool Exec', type: 'service', status: 'idle', x: 20, y: 70 },
                { id: 'n4', label: 'Synthesizer', type: 'agent', status: 'active', x: 80, y: 50 },
                { id: 'n5', label: 'Rogue Node', type: 'agent', status: 'stall', x: 80, y: 20 },
            ];

            const edges: SwarmEdge[] = [
                { source: 'n2', target: 'n1', status: 'healthy', hash: '0x1A4' },
                { source: 'n1', target: 'n3', status: 'healthy', hash: '0x2B9' },
                { source: 'n1', target: 'n4', status: 'healthy', hash: '0x3C1' },
                { source: 'n5', target: 'n1', status: 'blocked', hash: 'INVALID_GRAFT' },
            ];

            return {
                nodes,
                edges,
                anomalies: ['ARI Hub: Logic Graft Blocked from Rogue Node (n5)']
            };
        };

        const interval = setInterval(() => {
            // Simulate dynamic updates
            setData(generateMockData());
            setLoading(false);
        }, 3000);

        // Initial load
        setData(generateMockData());
        setLoading(false);"""

new_fetch = """        const fetchTopologyData = async () => {
            try {
                const response = await fetch('/api/v1/mock/swarm-topology');
                if (response.ok) {
                    const result = await response.json();
                    setData(result);
                }
            } catch (error) {
                console.error("Failed to fetch swarm topology data:", error);
            } finally {
                setLoading(false);
            }
        };

        const interval = setInterval(() => {
            fetchTopologyData();
        }, 3000);

        // Initial load
        fetchTopologyData();"""

content = content.replace(old_fetch, new_fetch)

with open(filename, 'w') as f:
    f.write(content)
