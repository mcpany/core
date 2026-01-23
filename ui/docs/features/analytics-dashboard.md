# Analytics Dashboard

The **Analytics Dashboard** provides real-time insights into your MCP infrastructure, allowing you to monitor traffic, performance, errors, and context usage across all your services.

![Dashboard Overview](../screenshots/dashboard_overview.png)

## Key Features

### Service Filtering
You can filter the entire dashboard by a specific **Service**. This allows you to isolate metrics for a single integration (e.g., "Postgres", "Stripe") to diagnose service-specific issues.

1.  Click the **"All Services"** dropdown in the top right.
2.  Select a service from the list.
3.  The charts (Request Volume, Latency, Errors) and "Top Tools" list will update to show data only for that service.

![Service Filter](../screenshots/dashboard_filtered.png)

### Performance Metrics
- **Total Requests**: The aggregate number of tool calls processed.
- **Avg Latency**: The average time taken to execute tools.
- **Error Rate**: The percentage of failed tool executions.

### Context Usage
The **Context Usage** tab visualizes the estimated token consumption of your tool definitions. This helps you identify "heavy" services that might be consuming too much context window in your LLM calls.

### Top Tools
Identify your most frequently used tools to understand agent behavior and optimize performance where it matters most.
