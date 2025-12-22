package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSettings_ExtraGetters(t *testing.T) {
	// Create a Settings instance manually with populated fields
	middlewares := []*configv1.Middleware{
		{
			Name: proto.String("test-middleware"),
		},
	}

	s := &Settings{
		dbPath: "/path/to/db.sqlite",
		proto: &configv1.GlobalSettings{
			Middlewares: middlewares,
		},
	}

	assert.Equal(t, "/path/to/db.sqlite", s.DBPath())
	assert.Equal(t, middlewares, s.Middlewares())
}
