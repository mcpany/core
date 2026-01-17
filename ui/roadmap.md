# UI Roadmap

## Status: Active Development

### Planned Features

- [ ] **Advanced Service Configuration & Sharing**:
  - [x] Visual editor for detailed service configuration (Connection, Auth, Advanced).
  - [x] Service Duplication: One-click cloning of existing services.
  - [x] Service Export: Download service configuration as JSON.
  - Import external services via gRPC auto-discovery or OpenAPI specs.
  - Export and share service configurations.
- [x] **Service Connection Diagnostic Tool**: Interactive tool to diagnose connection issues with upstream services (DNS, Handshake, Capabilities) - Added based on Ecosystem Audit.
  - _Update_: Enhanced with WebSocket support and browser-side connectivity probing.
- [ ] **Plugin UI Extensions**: Allow server plugins to inject custom UI components.
- [x] **Service Templates Library**: A built-in library of common service configurations (Postgres, Redis, Slack) to quickly spin up services without manual config.
- [ ] **Configuration Versioning & Rollback**: UI to view history of service configuration changes and rollback to previous versions.
- [ ] **Server Health History**: Visual timeline of server up/down status over the last 24h.
- [x] **Breadcrumb Navigation Enhancements**: Improved breadcrumbs with dropdowns for sibling navigation.
- [ ] **Drag-and-Drop Resource Export**: Ability to drag a resource from the list to the desktop or another app.
- [ ] **Resource Content Search**: Ability to search within the text content of resources for keywords.
- [x] **Binary Resource Preview**: Support for previewing images, PDFs, and other binary formats in the resource viewer.
- [ ] **Prompt/Resource Sibling Navigation**: Enable sibling navigation for Prompts and Resources (requires backend API update to include service_id).
- [ ] **Breadcrumb History**: Show recently visited breadcrumbs in a dropdown or history menu.
- [x] **JSON Schema Visualizer**: Display tool input schemas as interactive diagrams instead of raw JSON for better understanding of complex types.
- [ ] **Interactive Tool Usage History**: A timeline of tool executions with ability to replay them directly from the UI.
- [ ] **Bulk Service Import**: Allow importing multiple services from a single config file or URL.
- [x] **Service Configuration Validation**: Pre-save validation for service configs (e.g. check if URL is reachable).
- [x] **Service Tagging & Grouping**: Organize services by tags (e.g., prod, staging, external) and filter the list.
- [x] **Service Config Diff Viewer**: Visual diff when updating or duplicating services to see exactly what changed.
- [ ] **Bulk Service Actions**: Enable/Disable or Delete multiple services at once, potentially using tags for selection.
- [ ] **Service Config Diff Viewer**: Visual diff when updating or duplicating services to see exactly what changed.
- [x] **Bulk Service Actions**: Enable/Disable or Delete multiple services at once, potentially using tags for selection.
- [ ] **Tag-based Access Control**: Restrict service access to specific user profiles based on tags.
- [ ] **Live Tool Usage Graph**: Visual graph of tool execution metrics over time (RPS, Latency) in Tool Detail view.
- [ ] **Playground Schema Validation**: Enforce JSON schema validation in the Tool Playground before submission to prevent bad requests.
- [ ] **Favorites/Pinned Tools**: Ability to pin frequently used tools to the top of the list for quick access.
- [ ] **Schema Validation Playground**: Allow users to test JSON payloads against the schema directly in the inspector with validation feedback.
- [ ] **Service Health History Visualization**: A sparkline or small graph in the service list row showing uptime/latency history (e.g. last 1h).
- [ ] **Bulk Edit Configuration**: Ability to edit common properties (like tags, timeout, or environment variables) for multiple selected services.

### Completed Features

- [x] **JSON Schema Visualizer**: Implemented a recursive tree view for tool schemas, replacing raw JSON.
- [x] **Service Duplication & Export**: Added "Duplicate" and "Export" actions to the service list for easier management.

- [x] **Resource Preview Modal**: A dedicated modal to preview resources for larger viewing area.
- [x] **Service Template Integration**: Integrated Service Templates into the main "Add Service" flow for better discoverability.
- [x] **Context Menu for Resources**: Right-click interactions for resources (Copy URI, View Details) to improve usability.
- [x] **Global Keyboard Shortcuts Manager**: A dedicated UI to view and customize keyboard shortcuts for power users.
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
- [x] **System Status Banner**: A global banner that displays system health status and connectivity issues (polled from `/doctor`).
