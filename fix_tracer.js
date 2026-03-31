const fs = require('fs');

let content = fs.readFileSync('ui/src/components/dashboard/agent-chain-tracer.tsx', 'utf8');

// Make sure we don't have double useTraces
content = content.replace(/export function AgentChainTracer\(\) \{[\s\S]*?(?=return \()/m, `export function AgentChainTracer() {
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const { traces } = useTraces({ limit: 10 });

  const toggleExpand = (id: string) => {
    setExpandedId(expandedId === id ? null : id);
  };

  const chainData = traces.slice(0, 5).map((trace) => {
    let status = "active";
    if (trace.status === "success") status = "attested";
    else if (trace.status === "error") status = "speculative";

    let details = "No details provided.";
    if (trace.rootSpan?.errorMessage) {
      details = trace.rootSpan.errorMessage;
    } else if (trace.rootSpan?.input) {
      try {
        details = typeof trace.rootSpan.input === "string" ? trace.rootSpan.input : JSON.stringify(trace.rootSpan.input);
      } catch (e) {}
    }

    return {
      id: trace.id,
      agent: trace.rootSpan?.serviceName || trace.rootSpan?.name || "Unknown-Agent",
      action: trace.rootSpan?.name || "Unknown Action",
      status,
      latency: \`\${trace.totalDuration || 0}ms\`,
      hash: trace.id ? trace.id.substring(0, 12) : "0x000",
      details,
      timestamp: trace.timestamp ? format(new Date(trace.timestamp), "HH:mm:ss.SSS") : ""
    };
  });

  `);

fs.writeFileSync('ui/src/components/dashboard/agent-chain-tracer.tsx', content);
