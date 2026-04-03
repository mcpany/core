// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package doctor

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// PrintResults serves as a public interface for interacting with PrintResults.
//
// Summary: Print the results appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func PrintResults(w io.Writer, results []CheckResult) {
	if w == nil {
		w = os.Stdout
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	// No header for cleaner look, or maybe "STATUS\tSERVICE\tMESSAGE"

	for _, res := range results {
		var icon string
		switch res.Status {
		case StatusOk:
			icon = "✅"
		case StatusWarning:
			icon = "⚠️ "
		case StatusError:
			icon = "❌"
		case StatusSkipped:
			icon = "⏭️ "
		default:
			icon = "?"
		}

		_, _ = fmt.Fprintf(tw, "%s\t[%s]\t%s\t: %s\n", icon, res.Status, res.ServiceName, res.Message)
	}
	_ = tw.Flush()
}
