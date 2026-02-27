// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// visitedPool reuses maps to reduce allocations in EstimateJSONSize.
var jsonSizeVisitedPool = sync.Pool{
	New: func() interface{} {
		return make(map[uintptr]bool)
	},
}

// EstimateJSONSize estimates the size of the JSON representation of a value.
// It avoids allocating the full JSON string by traversing the structure recursively.
// It supports standard Go types and respects basic JSON encoding rules.
//
// Parameters:
//   - v (interface{}): The value to estimate.
//
// Returns:
//   - int: The estimated size in bytes.
func EstimateJSONSize(v interface{}) int {
	visited := jsonSizeVisitedPool.Get().(map[uintptr]bool)
	size := estimateJSONSizeRecursive(v, visited)

	// Clear map for reuse
	for k := range visited {
		delete(visited, k)
	}
	jsonSizeVisitedPool.Put(visited)
	return size
}

func estimateJSONSizeRecursive(v interface{}, visited map[uintptr]bool) int {
	if v == nil {
		return 4 // "null"
	}

	switch val := v.(type) {
	case string:
		return len(val) + 2
	case int:
		return estimateIntSize(int64(val))
	case int8:
		return estimateIntSize(int64(val))
	case int16:
		return estimateIntSize(int64(val))
	case int32:
		return estimateIntSize(int64(val))
	case int64:
		return estimateIntSize(val)
	case uint:
		return estimateUintSize(uint64(val))
	case uint8:
		return estimateUintSize(uint64(val))
	case uint16:
		return estimateUintSize(uint64(val))
	case uint32:
		return estimateUintSize(uint64(val))
	case uint64:
		return estimateUintSize(val)
	case float32:
		var buf [32]byte
		b := strconv.AppendFloat(buf[:0], float64(val), 'g', -1, 32)
		return len(b)
	case float64:
		var buf [32]byte
		b := strconv.AppendFloat(buf[:0], val, 'g', -1, 64)
		return len(b)
	case bool:
		if val {
			return 4 // "true"
		}
		return 5 // "false"
	case []byte:
		n := len(val)
		if n == 0 {
			return 2 // ""
		}
		return ((n+2)/3)*4 + 2
	case map[string]interface{}:
		return estimateMapSize(val, visited)
	case []interface{}:
		return estimateSliceSize(val, visited)
	// Fast paths for common types to avoid reflection
	case []string:
		size := 1 // [
		for i, s := range val {
			if i > 0 {
				size++
			}
			size += len(s) + 2
		}
		size++ // ]
		return size
	case map[string]string:
		size := 1 // {
		count := 0
		for k, s := range val {
			if count > 0 {
				size++
			}
			size += len(k) + 2 + 1 + len(s) + 2
			count++
		}
		size++ // }
		return size
	default:
		return estimateReflect(val, visited)
	}
}

func estimateIntSize(n int64) int {
	if n == 0 {
		return 1
	}
	if n > 0 {
		if n < 10 {
			return 1
		}
		if n < 100 {
			return 2
		}
		if n < 1000 {
			return 3
		}
	}
	var buf [24]byte
	return len(strconv.AppendInt(buf[:0], n, 10))
}

func estimateUintSize(n uint64) int {
	if n == 0 {
		return 1
	}
	if n < 10 {
		return 1
	}
	if n < 100 {
		return 2
	}
	var buf [24]byte
	return len(strconv.AppendUint(buf[:0], n, 10))
}

func estimateMapSize(m map[string]interface{}, visited map[uintptr]bool) int {
	if len(m) == 0 {
		return 2 // {}
	}

	ptr := reflect.ValueOf(m).Pointer()
	if visited[ptr] {
		return 0
	}
	visited[ptr] = true
	defer delete(visited, ptr)

	size := 1 // {
	count := 0
	for k, v := range m {
		if count > 0 {
			size++ // ,
		}
		size += len(k) + 2 // "key"
		size++ // :
		size += estimateJSONSizeRecursive(v, visited)
		count++
	}
	size++ // }
	return size
}

func estimateSliceSize(s []interface{}, visited map[uintptr]bool) int {
	if len(s) == 0 {
		return 2 // []
	}

	size := 1 // [
	for i, v := range s {
		if i > 0 {
			size++ // ,
		}
		size += estimateJSONSizeRecursive(v, visited)
	}
	size++ // ]
	return size
}

func estimateReflect(v interface{}, visited map[uintptr]bool) int {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return 4 // null
		}
		ptr := val.Pointer()
		if visited[ptr] {
			return 0
		}
		visited[ptr] = true
		res := estimateJSONSizeRecursive(val.Elem().Interface(), visited)
		delete(visited, ptr)
		return res

	case reflect.Struct:
		return estimateStruct(val, visited)

	case reflect.Map:
		return estimateReflectMap(val, visited)

	case reflect.Slice, reflect.Array:
		return estimateReflectSlice(val, visited)

	case reflect.Interface:
		if val.IsNil() {
			return 4
		}
		return estimateJSONSizeRecursive(val.Elem().Interface(), visited)

	default:
		return len(fmt.Sprintf("%v", v)) + 2
	}
}

func estimateStruct(val reflect.Value, visited map[uintptr]bool) int {
	typ := val.Type()
	size := 1 // {
	count := 0

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if fieldType.PkgPath != "" {
			continue // unexported
		}

		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		name := fieldType.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				name = parts[0]
			}
			if len(parts) > 1 && parts[1] == "omitempty" && isEmptyValue(field) {
				continue
			}
		}

		if count > 0 {
			size++ // ,
		}

		size += len(name) + 2 // "key"
		size++ // :

		if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface {
			if field.IsNil() {
				size += 4
			} else {
				size += estimateJSONSizeRecursive(field.Interface(), visited)
			}
		} else {
			// Optimization: specialized calls for common types to avoid Interface() allocation if possible
			// But Interface() is cleaner.
			size += estimateJSONSizeRecursive(field.Interface(), visited)
		}
		count++
	}
	size++ // }
	return size
}

func estimateReflectMap(val reflect.Value, visited map[uintptr]bool) int {
	if val.IsNil() {
		return 4 // null
	}
	ptr := val.Pointer()
	if visited[ptr] {
		return 0
	}
	visited[ptr] = true
	defer delete(visited, ptr)

	if val.Len() == 0 {
		return 2
	}

	size := 1
	count := 0
	iter := val.MapRange()
	for iter.Next() {
		if count > 0 {
			size++
		}
		k := iter.Key()
		// Optimization: Avoid fmt.Sprintf for string keys
		if k.Kind() == reflect.String {
			size += k.Len() + 2
		} else {
			keyStr := fmt.Sprintf("%v", k.Interface())
			size += len(keyStr) + 2
		}
		size++ // :

		size += estimateJSONSizeRecursive(iter.Value().Interface(), visited)
		count++
	}
	size++
	return size
}

func estimateReflectSlice(val reflect.Value, visited map[uintptr]bool) int {
	if val.Kind() == reflect.Slice && val.IsNil() {
		return 4 // null
	}
    if val.Kind() == reflect.Slice {
        ptr := val.Pointer()
        if visited[ptr] {
            return 0
        }
        visited[ptr] = true
        defer delete(visited, ptr)
    }

	if val.Len() == 0 {
		return 2
	}

	size := 1
	for i := 0; i < val.Len(); i++ {
		if i > 0 {
			size++
		}
		size += estimateJSONSizeRecursive(val.Index(i).Interface(), visited)
	}
	size++
	return size
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
