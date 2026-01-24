package doctor

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintHTMLResults(t *testing.T) {
	results := []CheckResult{
		{ServiceName: "s1", Status: StatusOk, Message: "All good"},
		{ServiceName: "s2", Status: StatusError, Message: "Bad"},
	}
	var buf bytes.Buffer
	PrintHTMLResults(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "<html>") {
		t.Error("Expected HTML output")
	}
	if !strings.Contains(out, "s1") || !strings.Contains(out, "All good") {
		t.Error("Expected service s1 details")
	}
	if !strings.Contains(out, "class=\"ERROR\"") {
		t.Error("Expected error class")
	}
}
