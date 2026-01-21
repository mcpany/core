# Context Usage Estimator

The Context Usage Estimator provides users with an approximate token count for each tool definition. This helps in managing the LLM context window size, preventing "context bloat" where too many tools consume available tokens.

## Feature Overview

-   **Location:** Tool Detail View ("Usage & Costs" card).
-   **Calculation:** Based on a heuristic of 4 characters per token for the JSON representation of the tool (Name, Description, Input Schema).
-   **Visuals:**
    -   Displays the estimated token count.
    -   Color-coded values (Green < 500, Yellow < 1000, Red > 1000).
    -   Tooltip explaining the estimation method.

## Screenshot

![Context Usage Estimator](../../../.audit/ui/2026-01-21/context_usage_estimator.png)
