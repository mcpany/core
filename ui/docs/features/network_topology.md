# Network Topology & Policy Management

The **Network Topology** view provides a real-time, visual representation of your MCP infrastructure. It combines traffic metrics, service health, and security policies into a unified graph, styled for clarity and professional operations.

## Features

### Visual Network Graph

- **Nodes**:
  - **Client (User)**: Represents the entry point of requests (e.g., specific user or global traffic).
  - **Gateway**: The central MCP Router/Firewall that enforces policies.
  - **Services**: Upstream services (e.g., Weather, Payments) connected to the gateway.
  - **Tools**: Individual Functions/Tools exposed by services.
- **Edges**:
  - Animated lines representing traffic flow.
  - Labels showing **Requests per Second (req/s)**.
  - Color-coded status (Red for errors/blocks).

### Real-Time Metrics

Clicking on any node reveals a detailed **Metrics Panel**:

- **Traffic Volume**: Requests/sec, Total Requests.
- **Latency**: Average response time.
- **Data Transfer**: Request/Response sizes.
- **Cache Performance**: Hit rates and bandwidth saved.
- **Firewall Stats**: Allowed vs Blocked requests.

### Policy Builder

A step-by-step wizard to create security policies intuitively.

1.  **Define**: Name your policy and choose an action (Allow/Deny).
2.  **Scope**: Select **Source** (User/Service) and **Destination** (Service/Tool).
3.  **Review**: Verify the impact before applying.

## How to Use

1.  **Navigate** to the **Topology** page from the sidebar.
2.  **View Status**: Quickly identify error-prone services (marked Amber/Red).
3.  **Inspect Nodes**: Click a node to diagnose issues using the sidebar.
4.  **Create Policy**:
    - Click the **(+) Create Policy** button.
    - Follow the wizard steps to block unwanted traffic or rate-limit specific tools.

## Technical Details

- **Data Source**: Metrics are aggregated from Prometheus and the MCP Router internals.
- **Updates**: The graph refreshes every 5 seconds.
