1. **`ui/docs/features/tool_analytics.md`**: Verify UI features for tool analytics exist. It seems `ui/src/app/stats/page.tsx` and `ui/src/components/stats/analytics-dashboard.tsx` exist.
2. **`ui/docs/features/auth.md`**: Verify auth features (`/login`, `/users` routes). The `/login` and `/users` routes exist in `ui/src/App.tsx`.
3. **`ui/docs/features/log-search-highlighting.md`**: Verify Live Logs search highlighting exists. I will check the code in `ui/src/app/logs/page.tsx` and `ui/src/components/logs`. Highlighting functionality is in `ui/src/components/logs/log-stream.tsx`.
4. **`ui/docs/features/resources.md`**: Verify `/resources` route and list. It exists in `ui/src/App.tsx` and `ui/src/components/resources`.
5. **`ui/docs/features/webhooks.md`**: Verify `/webhooks` and `/settings/webhooks` routes. They exist in `ui/src/App.tsx`.
6. **`ui/docs/features/secrets.md`**: Verify `/secrets` route and secrets vault. It exists in `ui/src/App.tsx` and `ui/src/components/settings/secrets-manager.tsx`.
7. **`ui/docs/features/connection-diagnostics.md`**: Verify Connection Diagnostics dialog. It exists in `ui/src/components/diagnostics/connection-diagnostic.tsx`.
8. **`server/docs/features/dynamic_registration.md`**: Verify dynamic tool registration. OpenAPI, gRPC, and GraphQL are all handled in `server/pkg/upstream`.
9. **`ui/docs/features/middleware.md`**: Verify `/middleware` route. It exists in `ui/src/App.tsx` and `ui/src/app/middleware/page.tsx`.
10. **`server/docs/features/mcpctl.md`**: Verify `mcpctl validate` and `mcpctl doctor` commands. `mcpctl` exists in `server/cmd/mcpctl`, and it has `doctor.go` and `main.go` containing `validate` and `doctor` commands.

11. **Run tests and linting**: Run `make test` and `make lint`.
12. **Complete pre-commit steps**: I will use `pre_commit_instructions` to ensure proper testing, verification, review, and reflection are done.
13. **Submit PR**: I will generate the audit report into `pr_description.md` and submit it.
