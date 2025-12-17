# 📊 Monitoring

MCP Any exposes a Prometheus metrics endpoint on the address specified by the `--metrics-listen-address` flag. If this flag is not specified, the metrics endpoint is disabled.

## Available Metrics

- `mcpany_tools_list_total`: Total number of `tools/list` requests.
- `mcpany_tools_call_total`: Total number of `tools/call` requests.
- `mcpany_tools_call_latency`: Latency of `tools/call` requests.
- `mcpany_tools_call_errors`: Total number of failed `tools/call` requests.
- `mcpany_tool_<tool_name>_call_total`: Total number of calls for a specific tool.
- `mcpany_tool_<tool_name>_call_latency`: Latency of calls for a specific tool.
- `mcpany_tool_<tool_name>_call_errors`: Total number of failed calls for a specific tool.
- `mcpany_config_reload_total`: Total number of configuration reloads.
- `mcpany_config_reload_errors`: Total number of failed configuration reloads.
- `mcpany_grpc_connections_opened_total`: Total number of opened gRPC connections.
- `mcpany_grpc_connections_closed_total`: Total number of closed gRPC connections.
- `mcpany_grpc_rpc_started_total`: Total number of started gRPC RPCs.
- `mcpany_grpc_rpc_finished_total`: Total number of finished gRPC RPCs.
