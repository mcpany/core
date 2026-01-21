// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// ValidateMapKeys validates that all keys in the map exist in the protobuf message descriptor.
// It recursively validates nested messages.
func ValidateMapKeys(path string, m map[string]interface{}, md protoreflect.MessageDescriptor) error {
	// Skip validation for google.protobuf.Struct as it allows arbitrary keys
	if md.FullName() == "google.protobuf.Struct" {
		return nil
	}

	for key, val := range m {
		fd := findField(md, key)
		if fd == nil {
			return handleUnknownField(path, key, md)
		}

		// Recurse if it's a message
		if val == nil {
			continue
		}

		if err := validateField(path, key, val, fd); err != nil {
			return err
		}
	}
	return nil
}

func handleUnknownField(path, key string, md protoreflect.MessageDescriptor) error {
	// Key not found. Find suggestions.
	suggestion := suggestFixContextAware(key, md)
	fullPath := key
	if path != "" {
		fullPath = path + "." + key
	}
	if suggestion != "" {
		return fmt.Errorf("unknown field %q in %s\n\n%s", fullPath, md.Name(), suggestion)
	}
	return fmt.Errorf("unknown field %q in %s", fullPath, md.Name())
}

func validateField(path, key string, val interface{}, fd protoreflect.FieldDescriptor) error {
	// Handle list of messages
	if fd.IsList() && fd.Kind() == protoreflect.MessageKind {
		return validateListField(path, key, val, fd)
	}

	if fd.Kind() == protoreflect.MessageKind && !fd.IsMap() {
		// Single message
		if valMap, ok := val.(map[string]interface{}); ok {
			newPath := path + "." + key
			if path == "" {
				newPath = key
			}
			return ValidateMapKeys(newPath, valMap, fd.Message())
		}
	}

	// We ignore Map fields (IsMap()) because map keys are dynamic (string/int).
	// But map values (if message) should be validated?
	if fd.IsMap() && fd.MapValue().Kind() == protoreflect.MessageKind {
		if valMap, ok := val.(map[string]interface{}); ok {
			for k, v := range valMap {
				if vMap, ok := v.(map[string]interface{}); ok {
					newPath := fmt.Sprintf("%s.%s[%s]", path, key, k)
					if path == "" {
						newPath = fmt.Sprintf("%s[%s]", key, k)
					}
					if err := ValidateMapKeys(newPath, vMap, fd.MapValue().Message()); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func validateListField(path, key string, val interface{}, fd protoreflect.FieldDescriptor) error {
	// val should be a slice
	if slice, ok := val.([]interface{}); ok {
		for i, item := range slice {
			if itemMap, ok := item.(map[string]interface{}); ok {
				newPath := fmt.Sprintf("%s.%s[%d]", path, key, i)
				if path == "" {
					newPath = fmt.Sprintf("%s[%d]", key, i)
				}
				if err := ValidateMapKeys(newPath, itemMap, fd.Message()); err != nil {
					return err
				}
			}
		}
	}
	// If it's a map (converted from yaml map with indices), handle that too
	if valMap, ok := val.(map[string]interface{}); ok {
		// We need to iterate values, assuming keys are indices or arbitrary
		for k, v := range valMap {
			if vMap, ok := v.(map[string]interface{}); ok {
				newPath := fmt.Sprintf("%s.%s[%s]", path, key, k)
				if path == "" {
					newPath = fmt.Sprintf("%s[%s]", key, k)
				}
				if err := ValidateMapKeys(newPath, vMap, fd.Message()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// suggestFixContextAware finds the closest matching field name in the SPECIFIC message descriptor.
func suggestFixContextAware(unknownField string, md protoreflect.MessageDescriptor) string {
	// Check common aliases
	aliases := map[string]string{
		"url":       "address",
		"uri":       "address",
		"endpoint":  "address",
		"endpoints": "address",
		"host":      "address",
		"cmd":       "command",
		"args":      "arguments",
	}

	// Only apply alias if the alias target ACTUALLY exists in this message
	if target, ok := aliases[strings.ToLower(unknownField)]; ok {
		if findField(md, target) != nil {
			return fmt.Sprintf("Did you mean %q? (Common alias)", target)
		}
	}

	candidates := []string{}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		candidates = append(candidates, string(fd.Name()))
		if fd.JSONName() != string(fd.Name()) {
			candidates = append(candidates, fd.JSONName())
		}
	}

	bestMatch := ""
	minDist := 100

	for _, name := range candidates {
		dist := levenshtein(unknownField, name)
		if dist < minDist {
			minDist = dist
			bestMatch = name
		}
	}

	limit := len(unknownField) / 2
	if len(unknownField) <= 3 {
		limit = 1
	} else if limit < 3 {
		limit = 3
	}

	if minDist <= limit {
		return fmt.Sprintf("Did you mean %q?", bestMatch)
	}

	// If no close match, verify if the field is perhaps valid in one of the sub-messages?
	// This would be "advanced context awareness" (e.g. user forgot to nest).
	// But let's stick to strict validation for now.

	return ""
}
