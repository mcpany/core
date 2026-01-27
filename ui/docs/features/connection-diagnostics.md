# Connection Diagnostics

The **Connection Diagnostic Tool** helps users troubleshoot connectivity issues between MCP Any and upstream services. It provides a guided, step-by-step analysis of the connection path, from client-side checks to backend health verification.

## Key Features

-   **Multi-Stage Analysis**:
    1.  **Configuration Check**: Validates the service configuration (URL format, required fields).
    2.  **Browser Connectivity**: Checks if the service is reachable from the user's browser (useful for identifying local network issues).
    3.  **Backend Health**: Queries the MCP Any backend to see if it can reach the upstream service.
    4.  **Operational Verification**: Attempts to list tools and resources to ensure full functionality.

-   **Smart Heuristics**:
    -   **Missing Configuration Detection**: Identifies common configuration errors such as missing environment variables (e.g., `GITHUB_TOKEN`, `API_KEY`) or authentication failures.
    -   **Localhost/Docker Detection**: Automatically detects if a user is trying to connect to `localhost` from within a Docker container and suggests using `host.docker.internal`.
    -   **Context-Aware Error Suggestions**: Analyzes error messages (e.g., "fetch failed", "connection refused", "404") and provides actionable advice.

-   **One-Click Fixes**:
    -   **Fix Configuration**: When a configuration or authentication error is detected (like a missing token), a direct link is provided to open the Service Editor at the exact tab (Authentication or Connection) needed to resolve the issue.

-   **Visual Logs**: Displays a real-time log of the diagnostic process, which can be copied to the clipboard for support.

## Usage

1.  Navigate to the **Services** page.
2.  Click the **Status Icon** (or the "Troubleshoot" button) next to any service.
3.  Click **Start Diagnostics**.
4.  Follow the on-screen progress and review the **Diagnostic Result** card for suggestions.
5.  If a fix is available, click **Fix Configuration** to jump directly to the relevant settings.

## Screenshots

### Diagnostic Failure Analysis
![Diagnostic Failure](../screenshots/diagnostics_fix_suggestion.png)

### Quick Fix Workflow
When clicking **Fix Configuration**, the editor opens directly to the required setting:
![Fix Configuration](../screenshots/diagnostics_fix_sheet.png)
