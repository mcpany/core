package filesystem

import (
	"context"
)

type filesystemToolDef struct {
	Name        string
	Description string
	Input       map[string]interface{}
	Output      map[string]interface{}
	Handler     func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error)
}
