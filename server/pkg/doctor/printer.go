// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package doctor

import (
	"fmt"
	"html"
	"io"
	"os"
	"text/tabwriter"
	"time"
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

		latencyStr := ""
		if res.Latency > 0 {
			latencyStr = fmt.Sprintf(" (%s)", res.Latency)
		}
		_, _ = fmt.Fprintf(tw, "%s\t[%s]\t%s\t: %s%s\n", icon, res.Status, res.ServiceName, res.Message, latencyStr)
	}
	_ = tw.Flush()
}

// GenerateHTML generates an HTML report from the check results.
func GenerateHTML(results []CheckResult) string {
	out := `<!DOCTYPE html>
<html>
<head>
	<title>MCP Any Doctor Report</title>
	<style>
		body { font-family: sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
		.result { border: 1px solid #ddd; margin-bottom: 10px; padding: 10px; border-radius: 5px; }
		.OK { background-color: #e6fffa; border-color: #b2f5ea; }
		.WARNING { background-color: #fffaf0; border-color: #fbd38d; }
		.ERROR { background-color: #fff5f5; border-color: #feb2b2; }
		.SKIPPED { background-color: #f7fafc; border-color: #edf2f7; }
		.icon { margin-right: 10px; font-size: 1.2em; }
		.status { font-weight: bold; }
	</style>
</head>
<body>
	<h1>🩺 MCP Any Doctor Report</h1>
	<p>Generated at: ` + time.Now().Format(time.RFC1123) + `</p>
`
	for _, res := range results {
		icon := "❓"
		switch res.Status {
		case StatusOk:
			icon = "✅"
		case StatusWarning:
			icon = "⚠️"
		case StatusError:
			icon = "❌"
		case StatusSkipped:
			icon = "⏭️"
		}
		out += fmt.Sprintf(`<div class="result %s">
			<span class="icon">%s</span>
			<span class="status">[%s]</span>
			<strong>%s</strong>: %s
		</div>`, res.Status, icon, res.Status, html.EscapeString(res.ServiceName), html.EscapeString(res.Message))
	}
	out += `</body></html>`
	return out
}
