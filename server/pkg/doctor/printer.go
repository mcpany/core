// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package doctor

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// PrintResults prints the doctor check results in a structured table.
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

		fixMsg := ""
		if res.Fix != nil {
			fixMsg = fmt.Sprintf(" (🔧 Fix available: %s)", res.FixName)
		}
		_, _ = fmt.Fprintf(tw, "%s\t[%s]\t%s\t: %s%s\n", icon, res.Status, res.ServiceName, res.Message, fixMsg)
	}
	_ = tw.Flush()
}
