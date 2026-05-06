1. **Optimize `CallTool` Logging Delay:** Avoid heavy marshaling by removing redundant unmarshal requests during tool executions.
2. **Optimize `CallTool` Level:** Improve `LazyLogResult.LogValue()` by pushing large payloads into `DEBUG` log levels instead of `INFO` (thus saving thousands of nanoseconds for map outputs by short-circuiting them in `INFO`).
3. **Redundant Decoding Cleanup:** Use directly unmarshaled map buffers in executions (like `HTTPTool`, `MCPTool`, `LocalCommandTool` and `CommandTool`) since we already correctly populate and use pre-configured JSON iterative parsing with fastJSON.
4. **Pre-commit**: Complete standard verification using `pre_commit_instructions`.
5. **Submit**: Request approval via `submit`.
