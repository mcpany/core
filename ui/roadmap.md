# UI Roadmap

## Status: Active Development

### Planned Features

- [ ] **Advanced Service Configuration & Sharing**:
  - [x] Visual editor for detailed service configuration (Connection, Auth, Advanced).
  - Import external services via gRPC auto-discovery or OpenAPI specs.
  - Export and share service configurations.
- [x] **Service Connection Diagnostic Tool**: Interactive tool to diagnose connection issues with upstream services (DNS, Handshake, Capabilities) - Added based on Ecosystem Audit.
- [ ] **Plugin UI Extensions**: Allow server plugins to inject custom UI components.
- [x] **Service Templates Library**: A built-in library of common service configurations (Postgres, Redis, Slack) to quickly spin up services without manual config.
- [ ] **Configuration Versioning & Rollback**: UI to view history of service configuration changes and rollback to previous versions.
- [ ] **Global Keyboard Shortcuts Manager**: A dedicated UI to view and customize keyboard shortcuts for power users.
- [ ] **Server Health History**: Visual timeline of server up/down status over the last 24h.

### Completed Features

- [x] **Global Search & Action Palette**: Enhanced Command Palette (Cmd+K) with navigation, system actions (Reload, Copy URL), and context-aware actions (Restart Service, Copy Resource URI).
- [x] **Service Environment Variable Editor**: UI for managing environment variables and working directory for command-line services, with secret masking.
- [x] **Network Topology Visualization**: Interactive graph of the MCP ecosystem. [Docs](server/docs/features/dynamic-ui.md)
- [x] **Middleware Visualization**: Drag-and-drop pipeline management. [Docs](server/docs/features/middleware_visualization.md)
- [x] **Real-time Log Streaming**: Live audit logs with backend integration (WebSocket). [Docs](server/docs/features/log_streaming_ui.md)
- [x] **Granular Real-time Metrics**: Display RPS for individual tools, upstream services, middleware, and webhooks in the dashboard. [Docs](server/docs/monitoring.md)
- [x] **Theme Builder**: Dark/Light mode support. [Docs](server/docs/features/theme_builder.md)
- [x] **Mobile Optimization**: Complete responsiveness for all management pages. [Docs](server/docs/mobile-view.md)
- [x] **Service Management**: Enable/Disable and configure services.
- [x] **Tool Playground**: Test tools with auto-generated forms.
- [x] **Profile Management**: Create and switch between user profiles.
- [x] **Observability Dashboard**: Real-time metrics and system health.
