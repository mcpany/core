1. **Goal:** Apply a random, high-impact performance fix and annotate it with `⚡ BOLT:`.
2. **Analysis:** The `EvaluateCallPolicy` function in `server/pkg/tool/policy.go` compiles regular expressions dynamically using `regexp.MatchString` and `regexp.Match` for each policy rule on every tool call. Although there is a compiled version (`EvaluateCompiledCallPolicy`), `EvaluateCallPolicy` acts as a fallback or may be called directly in some places. `regexp.MatchString` involves locking and compiling the regex on every single invocation, which is `O(n)` compilation overhead. This is a CPU bottleneck.
   We can introduce a `sync.Map` to cache the compiled regular expressions in `EvaluateCallPolicy` just like it's done for `ShouldExport`. This avoids repetitive compilations.
3. Wait, is `EvaluateCallPolicy` even used?
    ```bash
    grep -rn "EvaluateCallPolicy" server/pkg/
    ```
    It is used only in `policy_test.go`. The rest of the codebase uses `EvaluateCompiledCallPolicy`.
4. Let's look for `regexp.Compile` or `regexp.Match` elsewhere.
    `server/pkg/util/secrets.go` does `regexp.Compile(secret.GetValidationRegex())`.
