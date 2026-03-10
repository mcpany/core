1. **Target Identification**: In Phase 1, I scanned the codebase for missing `useMemo` optimizations in React components that perform heavy filtering or mapping on large data arrays.
2. **Shortlist Generation**: I found the following candidates:
    - `ui/src/app/tools/page.tsx` (`filteredTools` and `groupedTools`)
    - `ui/src/components/prompts/prompt-workbench.tsx` (`filteredPrompts`)
    - `ui/src/components/resources/resource-explorer.tsx` (`filteredResources`)
    - `ui/src/components/users/user-list.tsx` (`filteredUsers` - already has useMemo)
    - `ui/src/components/playground/pro/tool-sidebar.tsx` (`filteredTools` - already has useMemo)
3. **Selection**: I will randomly select one of the top targets without `useMemo`. Let's select `ui/src/app/tools/page.tsx` to fix the `filteredTools` and `groupedTools` calculation without `useMemo`.
4. **Implementation**:
    - Wrap the `filteredTools` and `groupedTools` logic in `useMemo` in `ui/src/app/tools/page.tsx`.
    - Also need to import `useMemo` in `ui/src/app/tools/page.tsx`.
    - Add the required `// ⚡ BOLT:` annotation to explain the rationale.
5. **Pre-commit Checks**: Run `make lint` and `make test`.
6. **Submission**: Commit with the specific message format.
