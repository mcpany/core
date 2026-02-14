// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
)

func TestValidateParameter(t *testing.T) {
	tests := []struct {
		name    string
		schema  *configv1.ParameterSchema
		value   any
		wantErr bool
	}{
		{
			name: "required parameter missing",
			schema: func() *configv1.ParameterSchema {
				s := &configv1.ParameterSchema{}
				s.SetName("param")
				s.SetIsRequired(true)
				return s
			}(),
			value:   nil,
			wantErr: true,
		},
		{
			name: "optional parameter missing",
			schema: func() *configv1.ParameterSchema {
				s := &configv1.ParameterSchema{}
				s.SetName("param")
				s.SetIsRequired(false)
				return s
			}(),
			value:   nil,
			wantErr: false,
		},
		{
			name: "pattern valid",
			schema: func() *configv1.ParameterSchema {
				s := &configv1.ParameterSchema{}
				s.SetPattern("^[a-z]+$")
				return s
			}(),
			value:   "abc",
			wantErr: false,
		},
		{
			name: "pattern invalid",
			schema: func() *configv1.ParameterSchema {
				s := &configv1.ParameterSchema{}
				s.SetPattern("^[a-z]+$")
				return s
			}(),
			value:   "123",
			wantErr: true,
		},
		{
			name: "min length valid",
			schema: func() *configv1.ParameterSchema {
				s := &configv1.ParameterSchema{}
				s.SetMinLength(3)
				return s
			}(),
			value:   "abc",
			wantErr: false,
		},
		{
			name: "min length invalid",
			schema: func() *configv1.ParameterSchema {
				s := &configv1.ParameterSchema{}
				s.SetMinLength(3)
				return s
			}(),
			value:   "ab",
			wantErr: true,
		},
		{
			name: "max length valid",
			schema: func() *configv1.ParameterSchema {
				s := &configv1.ParameterSchema{}
				s.SetMaxLength(3)
				return s
			}(),
			value:   "abc",
			wantErr: false,
		},
		{
			name: "max length invalid",
			schema: func() *configv1.ParameterSchema {
				s := &configv1.ParameterSchema{}
				s.SetMaxLength(3)
				return s
			}(),
			value:   "abcd",
			wantErr: true,
		},
		{
			name: "minimum valid",
			schema: func() *configv1.ParameterSchema {
				s := &configv1.ParameterSchema{}
				s.SetMinimum(10)
				return s
			}(),
			value:   10,
			wantErr: false,
		},
		{
			name: "minimum invalid",
			schema: func() *configv1.ParameterSchema {
				s := &configv1.ParameterSchema{}
				s.SetMinimum(10)
				return s
			}(),
			value:   9,
			wantErr: true,
		},
		{
			name: "maximum valid",
			schema: func() *configv1.ParameterSchema {
				s := &configv1.ParameterSchema{}
				s.SetMaximum(10)
				return s
			}(),
			value:   10,
			wantErr: false,
		},
		{
			name: "maximum invalid",
			schema: func() *configv1.ParameterSchema {
				s := &configv1.ParameterSchema{}
				s.SetMaximum(10)
				return s
			}(),
			value:   11,
			wantErr: true,
		},
		{
			name: "enum valid",
			schema: func() *configv1.ParameterSchema {
				s := &configv1.ParameterSchema{}
				s.SetEnum([]string{"A", "B"})
				return s
			}(),
			value:   "A",
			wantErr: false,
		},
		{
			name: "enum invalid",
			schema: func() *configv1.ParameterSchema {
				s := &configv1.ParameterSchema{}
				s.SetEnum([]string{"A", "B"})
				return s
			}(),
			value:   "C",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateParameter(tt.schema, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
