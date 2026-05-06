1. **Optimize `dedupTools` in `reconcileServices`**:
   - The current code:
     ```go
     deduplicated = append([]*config_v1.ToolDefinition{t}, deduplicated...)
     ```
     is an O(N^2) operation because it prepends to a slice in a loop.
   - We will replace this with:
     ```go
     var temp []*config_v1.ToolDefinition
     for i := len(tools) - 1; i >= 0; i-- {
         t := tools[i]
         if !seen[t.GetName()] {
             seen[t.GetName()] = true
             temp = append(temp, t)
         }
     }
     // Reverse temp
     for i, j := 0, len(temp)-1; i < j; i, j = i+1, j-1 {
         temp[i], temp[j] = temp[j], temp[i]
     }
     cmd.SetTools(temp)
     ```
     This changes it to O(N).

2. **Optimize cache key generation in `getCacheKey`**:
   - The `jsoniter.ConfigCompatibleWithStandardLibrary` uses `json.Marshal(req.Arguments)` which is slow but sorts keys. Is there a way to avoid it or optimize it? Wait, we can use `FastJSON` if we create a `FastJSONSortKeys` config:
     ```go
     var jsonSortKeys = jsoniter.Config{
         EscapeHTML:             false,
         SortMapKeys:            true,
         ValidateJsonRawMessage: true,
     }.Froze()
     ```
     Then use `jsonSortKeys.Marshal(req.Arguments)` instead of `json.Marshal`. It's significantly faster than standard library or compatible standard library. Add `FastJSONSortKeys` to `util/json_marshal.go`.

3. **Optimize `CallTool` latency in `server.go`**:
   - For unstructured result fallback, avoid the costly `util.FastMarshal` inside `CallTool`. Instead, use `util.FastMarshalToString(result)` to get `text` directly. Wait, `FastMarshalToString` allocates internally. But we can avoid `jsonBytes` entirely and just use `text`!
   - Wait, `util.FastMarshalToString(result)` is still an allocation.
   - Actually, in `LazyLogResult.LogValue()`: it currently calls `util.RedactJSON(jsonBytes)`. `RedactJSON` is VERY slow. If the map is large (like the benchmark map), `RedactJSON` scans it. We can't disable redaction. But we can truncate it, or check the length. Wait, `summarizeCallToolResult` already truncates text to 512 bytes! Unstructured maps are formatted into text, which IS NOT truncated when returned?
   - Wait, `CallToolResult` has `Content: []mcp.Content{&mcp.TextContent{Text: text}}`. Then `LogValue` calls `summarizeCallToolResult(ctr)`. Then `summarizeCallToolResult` truncates the text to 512 bytes!
   - But wait! Before that, `LazyLogResult.LogValue()` does:
     ```go
		if len(r.JSONBytes) > 0 {
			return slog.StringValue(util.BytesToString(util.RedactJSON(r.JSONBytes)))
		}
		jsonBytes, _ := util.FastMarshal(v)
		return slog.StringValue(util.BytesToString(util.RedactJSON(jsonBytes)))
     ```
     BUT ONLY IF `r.FinalResult == nil` !
     Wait, in `CallTool`, `finalResult` is populated! So `LazyLogResult.LogValue()` WILL call `summarizeCallToolResult(r.FinalResult)` and bypass `RedactJSON`!
     So `util.FastMarshal` in `CallTool` is purely to generate `text` for the `TextContent`.
     So changing `FastMarshal` to `FastMarshalToString` inside `CallTool` is the best optimization.
   - Even better, if we look at `server.go` 1025:
     ```go
		if len(jsonBytes) == 0 && marshalErr == nil {
			jsonBytes, marshalErr = util.FastMarshal(result)
			if marshalErr == nil {
				text = util.BytesToString(jsonBytes)
			}
		}
     ```
     We can just do:
     ```go
		text, err := util.FastMarshalToString(result)
		if err != nil {
			text = util.ToString(result)
		}
     ```
     This avoids creating a new byte slice and avoids returning `jsonBytes` that isn't used!

4. **Complete Pre-Commit Steps**
   - Use `pre_commit_instructions` tool to make sure proper testing, verifications, reviews, and reflections are done.

5. **Submit the Change**
   - Submit the branch using git conventions and the `submit` tool.
