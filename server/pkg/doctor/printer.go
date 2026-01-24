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

		_, _ = fmt.Fprintf(tw, "%s\t[%s]\t%s\t: %s\n", icon, res.Status, res.ServiceName, res.Message)
	}
	_ = tw.Flush()
}

// PrintHTMLResults prints the doctor check results as an HTML report.
func PrintHTMLResults(w io.Writer, results []CheckResult) {
	if w == nil {
		w = os.Stdout
	}

	html := `<!DOCTYPE html>
<html>
<head>
	<title>MCP Any Doctor Report</title>
	<style>
		body { font-family: sans-serif; }
		table { border-collapse: collapse; width: 100%; }
		th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
		th { background-color: #f2f2f2; }
		.OK { color: green; }
		.WARNING { color: orange; }
		.ERROR { color: red; }
		.SKIPPED { color: gray; }
	</style>
</head>
<body>
	<h1>MCP Any Doctor Report</h1>
	<table>
		<tr>
			<th>Status</th>
			<th>Service</th>
			<th>Message</th>
		</tr>
`
	_, _ = io.WriteString(w, html)

	for _, res := range results {
		_, _ = fmt.Fprintf(w, "\t\t<tr>\n")
		_, _ = fmt.Fprintf(w, "\t\t\t<td class=\"%s\">%s</td>\n", res.Status, res.Status)
		_, _ = fmt.Fprintf(w, "\t\t\t<td>%s</td>\n", res.ServiceName)
		_, _ = fmt.Fprintf(w, "\t\t\t<td>%s</td>\n", res.Message)
		_, _ = fmt.Fprintf(w, "\t\t</tr>\n")
	}

	_, _ = io.WriteString(w, "\t</table>\n</body>\n</html>\n")
}
