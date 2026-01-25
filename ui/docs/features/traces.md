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

### 2. Filtering
You can filter traces to isolate specific issues:
- **Search**: Filter by tool name or trace ID using the search bar.
- **Status Filter**: Toggle between "All", "Success", and "Error" tabs to focus on failed requests.

![Trace Filtered](screenshots/traces_filtered.png)

### 3. Drill-Down from Dashboard
Identify failing tools on the **Dashboard** and click on them in the "Tool Failure Rates" widget. This will automatically navigate to the Trace Inspector, filtered by that tool and "Error" status, allowing for immediate root cause analysis.

### 4. Inspect Detail
Click on any row in the trace list to open the **Detail View**.
this view is split into tabs:
- **Request**: The JSON arguments sent to the tool.
- **Response**: The JSON output returned.
- **Timeline**: A waterfall view of the operation (if sub-spans exist).

![Trace Detail](screenshots/trace_detail.png)

### 5. Replay Trace
To quickly reproduce a bug or test a tool:
1. Open a trace detail.
2. Click the **"Replay in Playground"** button.
3. You will be redirected to the Playground with the tool and arguments pre-filled.
