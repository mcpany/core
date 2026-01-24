# Tool Output Diffing

MCP Any Playground now supports intelligent output diffing for debugging tool executions.

## Overview

When developing or debugging MCP tools, it is often useful to compare the output of a tool execution with a previous execution of the same tool with identical arguments. This helps identify non-deterministic behavior, side effects, or changes in the backend data.

## How it Works

1.  **Execute a Tool**: Run a tool in the Playground (e.g., `calculator.add {"a": 1, "b": 2}`).
2.  **Execute Again**: Run the *same* tool with the *same* arguments.
3.  **View Diff**: If the output differs from the previous run, the Playground will automatically detect the change and display a **Diff** toggle in the result card.

## Visual Interface

-   **Result View**: Shows the raw JSON output of the current execution.
-   **Diff View**: Highlights the differences between the current output and the previous output.
    -   **Red**: Lines removed (present in previous, missing in current).
    -   **Green**: Lines added (present in current, missing in previous).

![Tool Diff Screenshot](../screenshots/playground_diff.png)

## Use Cases

-   **Debugging Non-Determinism**: Quickly spot if an LLM-based tool or a dynamic API is returning different results for the same input.
-   **Verifying State Changes**: Check if a `list_items` call returns different results after running a `add_item` call (by running `list_items` before and after).
-   **Regression Testing**: Manually verify if a code change in the tool implementation caused an unexpected change in output.
