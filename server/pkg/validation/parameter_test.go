package validation_test

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestValidateParameter(t *testing.T) {
	tests := []struct {
		name    string
		schema  *configv1.ParameterSchema
		value   any
		wantErr string
	}{
		{
			name: "Valid String",
			schema: configv1.ParameterSchema_builder{
				Name: proto.String("param"),
				Type: configv1.ParameterType_STRING.Enum(),
			}.Build(),
			value: "valid",
		},
		{
			name: "Required Missing",
			schema: configv1.ParameterSchema_builder{
				Name:       proto.String("param"),
				IsRequired: proto.Bool(true),
			}.Build(),
			value:   nil,
			wantErr: `parameter "param" is required`,
		},
		{
			name: "Min Length Failure",
			schema: configv1.ParameterSchema_builder{
				Name:      proto.String("param"),
				Type:      configv1.ParameterType_STRING.Enum(),
				MinLength: proto.Int32(5),
			}.Build(),
			value:   "abc",
			wantErr: `parameter "param" length 3 is less than minimum length 5`,
		},
		{
			name: "Max Length Failure",
			schema: configv1.ParameterSchema_builder{
				Name:      proto.String("param"),
				Type:      configv1.ParameterType_STRING.Enum(),
				MaxLength: proto.Int32(3),
			}.Build(),
			value:   "abcd",
			wantErr: `parameter "param" length 4 exceeds maximum length 3`,
		},
		{
			name: "Pattern Failure",
			schema: configv1.ParameterSchema_builder{
				Name:    proto.String("param"),
				Type:    configv1.ParameterType_STRING.Enum(),
				Pattern: proto.String("^[a-z]+$"),
			}.Build(),
			value:   "123",
			wantErr: `parameter "param" value "123" does not match pattern "^[a-z]+$"`,
		},
		{
			name: "Enum Failure",
			schema: configv1.ParameterSchema_builder{
				Name: proto.String("param"),
				Type: configv1.ParameterType_STRING.Enum(),
				Enum: []string{"A", "B"},
			}.Build(),
			value:   "C",
			wantErr: `parameter "param" value "C" is not allowed; allowed values: [A B]`,
		},
		{
			name: "Minimum Failure",
			schema: configv1.ParameterSchema_builder{
				Name:    proto.String("param"),
				Type:    configv1.ParameterType_INTEGER.Enum(),
				Minimum: proto.Float64(10),
			}.Build(),
			value:   5,
			wantErr: `parameter "param" value 5 is less than minimum 10`,
		},
		{
			name: "Maximum Failure",
			schema: configv1.ParameterSchema_builder{
				Name:    proto.String("param"),
				Type:    configv1.ParameterType_NUMBER.Enum(),
				Maximum: proto.Float64(100.5),
			}.Build(),
			value:   100.6,
			wantErr: `parameter "param" value 100.6 exceeds maximum 100.5`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validation.ValidateParameter(tt.schema, tt.value)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
