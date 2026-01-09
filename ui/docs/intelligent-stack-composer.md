# Intelligent Stack Composer

The **Intelligent Stack Composer** transforms the raw "Config-as-Code" experience into a polished, Apple-style management console. It bridges the gap between low-level YAML editing and high-level visual composition.

## Features

1.  **Service Palette**: A sidebar providing drag-and-drop access to common service templates (Postgres, Redis, MCP Servers, etc.).
2.  **Intelligent Editor**: A robust YAML editor with syntax highlighting, line numbers, and real-time validation.
3.  **Live Visualizer**: A real-time preview of the stack topology, parsing the YAML configuration into a visual representation of services, their types, and configurations.
4.  **Error Handling**: Immediate feedback on YAML syntax errors with detailed messages.

## Usage

1.  Navigate to the **Stack Editor** (e.g., via Stacks menu).
2.  Use the **Service Palette** on the left to browse available templates.
3.  Click the **+** icon on a template to inject its configuration into your stack.
4.  Edit the YAML manually in the center pane.
5.  Observe the **Live Preview** on the right updating in real-time to reflect your changes.
6.  Click **Save Changes** to persist the configuration.

## Technical Details

-   **Parsing**: Uses `js-yaml` for safe client-side parsing.
-   **Validation**: Real-time validation loop provides instant feedback.
-   **Architecture**: 3-pane layout with toggleable sidebars for maximum screen real estate.
