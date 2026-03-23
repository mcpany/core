import React from 'react';

export default function UniversalAgentBusPage() {
  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold mb-4">Universal Agent Bus</h1>
      <p className="text-gray-600 mb-8">
        Visualize, configure, and orchestrate interactions between complex AI swarms.
      </p>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="border rounded-lg p-6 bg-white shadow-sm">
          <h2 className="text-xl font-semibold mb-4">Agent Chain Tracer</h2>
          <div className="text-center p-8 text-gray-500 border border-dashed rounded bg-gray-50">
            No multi-agent handoffs currently active.
          </div>
        </div>

        <div className="border rounded-lg p-6 bg-white shadow-sm">
          <h2 className="text-xl font-semibold mb-4">Unified Discovery Manager</h2>
          <div className="text-center p-8 text-gray-500 border border-dashed rounded bg-gray-50">
            No MCP servers auto-discovered.
          </div>
        </div>
      </div>
    </div>
  );
}
