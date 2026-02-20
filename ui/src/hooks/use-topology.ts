/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useEffect, useMemo } from 'react';
import { useTopology as useTopologyContext } from '../contexts/service-health-context';
import { transformGraphToReactFlow, GraphResponse } from '../components/visualizer/graph-utils';

/**
 * useTopology hook to get React Flow compatible nodes and edges from the system topology.
 * It uses the global ServiceHealthContext for data polling.
 */
export function useTopology() {
    const { latestTopology, refreshTopology } = useTopologyContext();

    const { nodes, edges } = useMemo(() => {
        // Cast types/topology Graph to graph-utils GraphResponse (structure is compatible)
        return transformGraphToReactFlow(latestTopology as unknown as GraphResponse);
    }, [latestTopology]);

    useEffect(() => {
        // Trigger a refresh on mount to ensure fresh data
        refreshTopology();
    }, [refreshTopology]);

    return {
        nodes,
        edges,
        refresh: refreshTopology,
        loading: !latestTopology
    };
}
