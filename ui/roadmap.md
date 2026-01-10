# UI Roadmap

## Status: Active Development

### Planned Features

- [ ] **Advanced Service Configuration & Sharing**:
  - UI-based generation, editing, and management of upstream service configurations.
  - Import external services via gRPC auto-discovery or OpenAPI specs.
  - Visual editor to enable/disable parts of the spec, modify defaults, add fields, and link secrets to local authentication variables.
  - Export and share service configurations.
- [ ] **Plugin UI Extensions**: Allow server plugins to inject custom UI components.

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
- [x] **Traffic Inspector**: Live inspection of HTTP traffic with details view.
