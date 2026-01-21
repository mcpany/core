package tool

import (
	"testing"
)

func TestCheckForShellInjection_EmptyTemplate_Space(t *testing.T) {
	val := "hello world"
	template := ""
	placeholder := ""
	command := "env"

	err := checkForShellInjection(val, template, placeholder, command)
	if err != nil {
		t.Fatalf("Unexpected error for space with empty template: %v. charsToCheck likely contains space.", err)
	}
}
