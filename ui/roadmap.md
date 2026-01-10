# UI Roadmap

## Status: Active Development

### Planned Features

- [ ] **Advanced Service Configuration & Sharing**:
  - Import external services via gRPC auto-discovery or OpenAPI specs.
  - Visual editor to enable/disable parts of the spec, modify defaults, add fields.
  - Export and share service configurations.
- [ ] **Plugin UI Extensions**: Allow server plugins to inject custom UI components.
- [ ] **Service Templates Library**: A built-in library of common service configurations (Postgres, Redis, Slack) to quickly spin up services without manual config.
- [ ] **Configuration Versioning & Rollback**: UI to view history of service configuration changes and rollback to previous versions.

### Completed Features

- [x] **Service Environment Variable Editor**: UI for managing environment variables and working directory for command-line services, with secret masking.
- [x] **Network Topology Visualization**: Interactive graph of the MCP ecosystem. [Docs](server/docs/features/dynamic-ui.md)

### Completed Features

- [x] **Network Topology Visualization**: Interactive graph of the MCP ecosystem. [Docs](server/docs/features/dynamic-ui.md)
- [x] **Middleware Visualization**: Drag-and-drop pipeline management. [Docs](server/docs/features/middleware_visualization.md)
- [x] **Real-time Log Streaming UI**: Live audit logs. [Docs](server/docs/features/log_streaming_ui.md)
- [x] **Granular Real-time Metrics**: Display RPS for individual tools, upstream services, middleware, and webhooks in the dashboard. [Docs](server/docs/monitoring.md)
- [x] **Theme Builder**: Dark/Light mode support. [Docs](server/docs/features/theme_builder.md)
- [x] **Mobile Optimization**: Complete responsiveness for all management pages. [Docs](server/docs/mobile-view.md)
- [x] **Service Management**: Enable/Disable and configure services.
- [x] **Tool Playground**: Test tools with auto-generated forms.
- [x] **Profile Management**: Create and switch between user profiles.
- [x] **Observability Dashboard**: Real-time metrics and system health.
