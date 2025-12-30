# Tool Inspector Feature

Today, I have chosen to build: **Tool Inspector & Tester** because developers need a way to inspect tool schemas and test tool execution in isolation without relying on chat-based interactions.

## Features

1.  **Schema Visualization**: View the JSON schema of any tool input arguments.
2.  **Interactive Testing**: A dynamic form generated from the schema allows users to input arguments.
3.  **Execution Feedback**: Execute the tool and see the structured JSON result immediately.
4.  **Seamless Integration**: Opens as a side drawer (Sheet) from the Tools list, maintaining context.

## Implementation Details

-   **Component**: `ToolInspector` in `ui/src/components/tools/tool-inspector.tsx`.
-   **Integration**: Added "Inspect" button to `ui/src/app/tools/page.tsx`.
-   **Backend**: Enhanced `ui/src/app/api/tools/route.ts` to support mock execution and return schemas.

## Screenshot

![Tool Inspector](tool_inspector.png)
