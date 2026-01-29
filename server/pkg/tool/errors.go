package tool

import "errors"

// ErrToolNotFound is returned when a requested tool cannot be found.
var ErrToolNotFound = errors.New("unknown tool")
