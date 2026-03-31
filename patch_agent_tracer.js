const fs = require('fs');
let content = fs.readFileSync('ui/src/components/dashboard/agent-chain-tracer.tsx', 'utf8');

// Replace the mock data
content = content.replace(/const MOCK_CHAIN_DATA = \[[\s\S]*?\];/g, `import { useTraces } from "@/hooks/use-traces";
import { format } from "date-fns";`);

content = content.replace(/export function AgentChainTracer\(\) \{/g, `export function AgentChainTracer() {
  const { traces } = useTraces({ limit: 10 });

  const chainData = traces.slice(0, 5).map((trace) => {
    let status = "active";
    if (trace.status === "success") status = "attested";
    else if (trace.status === "error") status = "speculative";

    let details = "No details provided.";
    if (trace.rootSpan.errorMessage) {
      details = trace.rootSpan.errorMessage;
    } else if (trace.rootSpan.input) {
      try {
        details = typeof trace.rootSpan.input === "string" ? trace.rootSpan.input : JSON.stringify(trace.rootSpan.input);
      } catch (e) {}
    }

    return {
      id: trace.id,
      agent: trace.rootSpan.serviceName || trace.rootSpan.name || "Unknown-Agent",
      action: trace.rootSpan.name || "Unknown Action",
      status,
      latency: \`\${trace.totalDuration}ms\`,
      hash: trace.id.substring(0, 12),
      details,
      timestamp: format(new Date(trace.timestamp), "HH:mm:ss.SSS")
    };
  });
`);

content = content.replace(/MOCK_CHAIN_DATA/g, 'chainData');

fs.writeFileSync('ui/src/components/dashboard/agent-chain-tracer.tsx', content);
