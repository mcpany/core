package util

import (
	"bytes"
	"testing"
)

func TestRedactMissingPlurals(t *testing.T) {
	// The bug was that "credentials" (plural) was not in the sensitive list,
	// and the boundary check logic prevented "credential" from matching "credentials".
	// Similarly for "secrets".
	input := `{"credentials": "very secret", "secrets": "super secret", "credential": "redacted", "secret": "redacted"}`

	got := RedactJSON([]byte(input))

    if bytes.Contains(got, []byte("very secret")) {
         t.Errorf("credentials value was not redacted: %s", string(got))
    }

	if bytes.Contains(got, []byte("super secret")) {
		t.Errorf("secrets value was not redacted: %s", string(got))
    }

	// Ensure singular forms are also redacted (regression check)
	// We look for the literal string "redacted" which was the value.
	// It should be replaced by "[REDACTED]".
	if bytes.Contains(got, []byte(`"redacted"`)) {
		t.Errorf("singular credential/secret value was not redacted: %s", string(got))
	}
}
