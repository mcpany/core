# UI Roadmap

## Status: Active Development

### Planned Features

- [ ] **Advanced Middleware Visualization**: Interactive editing of middleware chains directly from the graph view.
- [ ] **Granular Real-time Metrics**: Display RPS for individual tools, upstream services, middleware, and webhooks in the dashboard.
- [ ] **Advanced Service Configuration & Sharing**:
  - UI-based generation, editing, and management of upstream service configurations.
  - Import external services via gRPC auto-discovery or OpenAPI specs.
  - Visual editor to enable/disable parts of the spec, modify defaults, add fields, and link secrets to local authentication variables.
  - Export and share service configurations.
- [ ] **Real-time Log Streaming**: Enhanced filtering and search capabilities for high-volume log streams.
- [ ] **Mobile Optimization**: Complete responsiveness for all management pages.
- [ ] **Theme Builder**: Custom color themes beyond the built-in Light/Dark modes.
- [ ] **Plugin UI Extensions**: Allow server plugins to inject custom UI components.

## Completed Features

- [x] **Network Topology Visualization**: Interactive graph of the MCP ecosystem.
- [x] **Service Management**: Enable/Disable and configure services.
- [x] **Tool Playground**: Test tools with auto-generated forms.
- [x] **Profile Management**: Create and switch between user profiles.
- [x] **Observability Dashboard**: Real-time metrics and system health.
