// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// ValidateParameter checks if the value conforms to the schema constraints.
//
// Parameters:
//   - schema: *configv1.ParameterSchema. The schema to validate against.
//   - value: any. The value to validate.
//
// Returns:
//   - error: An error if validation fails.
func ValidateParameter(schema *configv1.ParameterSchema, value any) error {
	if schema == nil {
		return nil
	}

	if value == nil {
		if schema.GetIsRequired() {
			return fmt.Errorf("parameter %q is required", schema.GetName())
		}
		return nil
	}

	valStr := toString(value)

	// Enum Validation
	if len(schema.GetEnum()) > 0 {
		allowed := false
		for _, v := range schema.GetEnum() {
			if v == valStr {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("parameter %q value %q is not allowed; allowed values: %v", schema.GetName(), valStr, schema.GetEnum())
		}
	}

	// String validations
	if schema.GetType() == configv1.ParameterType_STRING {
		if schema.HasMinLength() {
			if int32(len(valStr)) < schema.GetMinLength() {
				return fmt.Errorf("parameter %q length %d is less than minimum length %d", schema.GetName(), len(valStr), schema.GetMinLength())
			}
		}
		if schema.HasMaxLength() {
			if int32(len(valStr)) > schema.GetMaxLength() {
				return fmt.Errorf("parameter %q length %d exceeds maximum length %d", schema.GetName(), len(valStr), schema.GetMaxLength())
			}
		}
		if schema.HasPattern() {
			matched, err := regexp.MatchString(schema.GetPattern(), valStr)
			if err != nil {
				return fmt.Errorf("parameter %q has invalid regex pattern: %w", schema.GetName(), err)
			}
			if !matched {
				return fmt.Errorf("parameter %q value %q does not match pattern %q", schema.GetName(), valStr, schema.GetPattern())
			}
		}
	}

	// Number validations
	if schema.GetType() == configv1.ParameterType_NUMBER || schema.GetType() == configv1.ParameterType_INTEGER {
		valFloat, err := convertToFloat(value)
		if err != nil {
			// If it's not a number but type says number, we should technically fail.
			// But util.ToString converts it to string.
			// If we can't parse it as float, it's definitely invalid for NUMBER.
			return fmt.Errorf("parameter %q expected number, got %v", schema.GetName(), value)
		}

		if schema.HasMinimum() {
			if valFloat < schema.GetMinimum() {
				return fmt.Errorf("parameter %q value %v is less than minimum %v", schema.GetName(), valFloat, schema.GetMinimum())
			}
		}
		if schema.HasMaximum() {
			if valFloat > schema.GetMaximum() {
				return fmt.Errorf("parameter %q value %v exceeds maximum %v", schema.GetName(), valFloat, schema.GetMaximum())
			}
		}
	}

	return nil
}

func convertToFloat(v any) (float64, error) {
	switch val := v.(type) {
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	case json.Number:
		return val.Float64()
	}
	return 0, fmt.Errorf("not a number")
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case json.Number:
		return val.String()
	case fmt.Stringer:
		return val.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
