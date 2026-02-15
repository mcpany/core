package doctor

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// PrintResults prints the doctor check results in a structured table to the provided writer.
//
// It formats the check results with status icons and alignment for readability.
//
// Parameters:
//   - w: io.Writer. The writer to output the results to (e.g., os.Stdout). If nil, defaults to os.Stdout.
//   - results: []CheckResult. The list of check results to print.
//
// Returns:
//   None.
//
// Side Effects:
//   - Writes formatted text to the provided writer.
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
