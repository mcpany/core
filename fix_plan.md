# Verification Plan

1. **`ui/docs/features/tool_analytics.md`**: Ensure Tool Analytics UI features are present.
2. **`ui/docs/features/auth.md`**: Ensure basic auth features (`/login`, `/users` routes) exist.
3. **`ui/docs/features/log-search-highlighting.md`**: Check if live logs search highlighting is in the UI.
4. **`ui/docs/features/resources.md`**: Check if `/resources` route and list exist.
5. **`ui/docs/features/webhooks.md`**: Check if `/webhooks` route and webhook features exist.
6. **`ui/docs/features/secrets.md`**: Check if `/secrets` route and secrets vault exist.
7. **`ui/docs/features/connection-diagnostics.md`**: Check if Connection Diagnostics dialog exists.
8. **`server/docs/features/dynamic_registration.md`**: Verify server code for dynamic tool registration (OpenAPI, gRPC, GraphQL).
9. **`ui/docs/features/middleware.md`**: Check if `/middleware` route exists.
10. **`server/docs/features/mcpctl.md`**: Verify `mcpctl validate` and `mcpctl doctor` commands exist in server.

## Remediation Plan

I will verify each of these 10 items.
For UI items (1-7, 9), I will check `ui/src/App.tsx` and the corresponding component files.
For Server items (8, 10), I will check the `server/cmd/mcpctl` and `server/pkg/` code.

If code is missing for a documented feature (Roadmap Debt), I will implement it.
If documentation is outdated compared to existing code (Documentation Drift), I will update the documentation.
