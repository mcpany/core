package tool

import (
	"errors"
	"testing"
)

func TestErrToolNotFound(t *testing.T) {
	err := ErrToolNotFound
	if err == nil {
		t.Error("Expected ErrToolNotFound to be non-nil")
	}

	if !errors.Is(err, ErrToolNotFound) {
		t.Errorf("Expected error to be ErrToolNotFound, got %v", err)
	}
}
