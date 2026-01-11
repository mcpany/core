# Live Trace Inspector

**Status:** Implemented

## Goal
Debug complex interactions by inspecting the full lifecycle of MCP requests. The Trace Inspector allows you to examine payloads, timing, and errors for every tool call and API request.

## Usage Guide

### 1. Trace List
Navigate to `/traces`. This view shows a chronological log of all system activity.

![Trace List](screenshots/traces_list.png)

- **Status Icons**: Green check for success, Red X for failure.
- **Duration**: Time taken for the request to complete.

### 2. Inspect Detail
Click on any row in the trace list to open the **Detail View**.
this view is split into tabs:
- **Request**: The JSON arguments sent to the tool.
- **Response**: The JSON output returned.
- **Timeline**: A waterfall view of the operation (if sub-spans exist).

![Trace Detail](screenshots/trace_detail.png)

### 3. Replay Trace
To quickly reproduce a bug or test a tool:
1. Open a trace detail.
2. Click the **"Replay in Playground"** button.
3. You will be redirected to the Playground with the tool and arguments pre-filled.
