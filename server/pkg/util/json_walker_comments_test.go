package util

import (
	"testing"
)

func TestWalkJSONStrings_Bug_CommentsWithPrecedingSlash(t *testing.T) {
	// Logic bug: WalkJSONStrings only checks the first slash in the segment preceding a quote.
	// If that slash is not a comment start (e.g. division operator), but a later slash IS a comment start,
	// and that comment contains a quote, WalkJSONStrings will incorrectly identify the quote inside the comment as a real string.

	input := []byte(`{"a": 10 / 2 /* "commented" */, "b": "value"}`)

	var visited []string
	WalkJSONStrings(input, func(raw []byte) ([]byte, bool) {
		visited = append(visited, string(raw))
		return nil, false
	})

	// Expected: "value" (and maybe "a", "b" depending on if they are keys? WalkJSONStrings visits values, not keys.)
	// "a" and "b" are keys.
	// So we expect ONLY "value".

	// If bug exists, it will also visit "commented".

	expected := []string{`"value"`}

	if len(visited) != len(expected) {
		t.Errorf("Expected visited %v, got %v", expected, visited)
		return
	}

	for i := range visited {
		if visited[i] != expected[i] {
			t.Errorf("Expected visited[%d] = %s, got %s", i, expected[i], visited[i])
		}
	}
}
