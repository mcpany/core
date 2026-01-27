# Tool Execution Replay

The Tool Inspector now includes a powerful **Replay** feature that allows users to quickly re-run previous tool executions. This is essential for debugging failed calls or iteratively testing tool behavior with slight modifications to arguments.

## Features

- **Execution History**: View a timeline of recent tool executions, including status, latency, and arguments.
- **One-Click Replay**: Click the **Replay** button (counter-clockwise arrow) on any history entry to load its arguments into the editor.
- **Auto-Switching**: The interface automatically switches to the "Test & Execute" tab so you can immediately run the tool.
- **Arguments Preview**: See a snapshot of the arguments used in each call directly in the timeline.

## Usage

1. Navigate to the **Tools** page.
2. Click **Inspect** on any tool.
3. Switch to the **Performance & Analytics** tab.
4. Locate a previous execution in the **Recent Timeline**.
5. Click the **Replay** icon button.
6. The arguments will be loaded into the input area, ready for execution.

![Tool Replay](../screenshots/tool_replay.png)
