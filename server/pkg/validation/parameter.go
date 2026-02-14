// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// ValidateParameter validates a value against a parameter schema.
//
// Summary: Validates a value against a parameter schema.
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

	if isNil(value) {
		if schema.GetIsRequired() {
			return fmt.Errorf("missing required parameter: %s", schema.GetName())
		}
		return nil
	}

	valStr := toString(value)

	// 1. Pattern (String only)
	if schema.HasPattern() {
		matched, err := regexp.MatchString(schema.GetPattern(), valStr)
		if err != nil {
			return fmt.Errorf("invalid regex pattern %q: %w", schema.GetPattern(), err)
		}
		if !matched {
			return fmt.Errorf("value does not match pattern %q", schema.GetPattern())
		}
	}

	// 2. Min/Max Length (String only)
	if schema.HasMinLength() {
		if len(valStr) < int(schema.GetMinLength()) {
			return fmt.Errorf("value length %d is less than minimum %d", len(valStr), schema.GetMinLength())
		}
	}
	if schema.HasMaxLength() {
		if len(valStr) > int(schema.GetMaxLength()) {
			return fmt.Errorf("value length %d is greater than maximum %d", len(valStr), schema.GetMaxLength())
		}
	}

	// 3. Minimum/Maximum (Number/Integer)
	if schema.HasMinimum() || schema.HasMaximum() {
		// Try to parse float. If it's not a number, we only error if numeric validation is requested.
		// If it's a string type but has numeric constraints, we try to parse it.
		f, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			// If validation requires numeric check but value is not numeric
			return fmt.Errorf("value is not a valid number for numeric validation: %v", err)
		}

		if schema.HasMinimum() {
			if f < schema.GetMinimum() {
				return fmt.Errorf("value %v is less than minimum %v", f, schema.GetMinimum())
			}
		}
		if schema.HasMaximum() {
			if f > schema.GetMaximum() {
				return fmt.Errorf("value %v is greater than maximum %v", f, schema.GetMaximum())
			}
		}
	}

	// 4. Enum (String)
	if len(schema.GetEnum()) > 0 {
		found := false
		for _, allowed := range schema.GetEnum() {
			if valStr == allowed {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value %q is not in allowed enum list: %v", valStr, schema.GetEnum())
		}
	}

	return nil
}

// Helper functions to avoid import cycle with util package

func isNil(i any) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Chan, reflect.Slice, reflect.Func, reflect.UnsafePointer:
		return v.IsNil()
	}
	return false
}

func toString(v any) string {
	if v == nil {
		return "<nil>"
	}
	switch val := v.(type) {
	case string:
		return val
	case json.Number:
		return val.String()
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(val)
	case int8:
		return strconv.FormatInt(int64(val), 10)
	case int16:
		return strconv.FormatInt(int64(val), 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case uint8:
		return strconv.FormatUint(uint64(val), 10)
	case uint16:
		return strconv.FormatUint(uint64(val), 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float32:
		if float64(val) == math.Trunc(float64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(float64(val), 'g', -1, 32)
	case float64:
		if math.Trunc(val) == val {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'g', -1, 64)
	case fmt.Stringer:
		return val.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
