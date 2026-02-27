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

// ⚡ Bolt: Struct field cache to avoid repetitive reflection calls.
// Randomized Selection from Top 5 High-Impact Targets
var structFieldCache sync.Map // map[reflect.Type][]cachedField

type cachedField struct {
	Index        int
	Name         string
	OmitEmpty    bool
	IsUnexported bool
}

func getCachedFields(typ reflect.Type) []cachedField {
	if v, ok := structFieldCache.Load(typ); ok {
		return v.([]cachedField)
	}

	var fields []cachedField
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		cf := cachedField{
			Index: i,
			Name:  field.Name,
		}

		if field.PkgPath != "" {
			cf.IsUnexported = true
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				cf.Name = parts[0]
			}
			if len(parts) > 1 && parts[1] == "omitempty" {
				cf.OmitEmpty = true
			}
		}
		fields = append(fields, cf)
	}
	structFieldCache.Store(typ, fields)
	return fields
}

// EstimateJSONSize estimates the size of the JSON representation of a value.
// It avoids allocating the full JSON string by traversing the structure recursively.
// It supports standard Go types and respects basic JSON encoding rules.
func EstimateJSONSize(v interface{}) int {
	// Optimization: Handle simple types without map allocation
	if size, handled := estimatePrimitiveSize(v); handled {
		return size
	}

	visited := jsonSizeVisitedPool.Get().(map[uintptr]bool)
	size := estimateJSONSizeRecursive(v, visited)

	// Clear map for reuse
	for k := range visited {
		delete(visited, k)
	}
	jsonSizeVisitedPool.Put(visited)
	return size
}

// estimatePrimitiveSize checks for primitive types that don't need recursion.
func estimatePrimitiveSize(v interface{}) (int, bool) {
	switch val := v.(type) {
	case nil:
		return 4, true
	case string:
		return len(val) + 2, true
	case int:
		return estimateIntSize(int64(val)), true
	case int64:
		return estimateIntSize(val), true
	case bool:
		if val {
			return 4, true
		}
		return 5, true
	case []byte:
		n := len(val)
		if n == 0 {
			return 2, true // ""
		}
		// Base64 estimation: 4 * ceil(n/3)
		return ((n+2)/3)*4 + 2, true
	}
	return 0, false
}

func estimateJSONSizeRecursive(v interface{}, visited map[uintptr]bool) int {
	// Try primitive check first (handles nil, string, int, bool, []byte)
	if size, handled := estimatePrimitiveSize(v); handled {
		return size
	}

	// Handle other types
	switch val := v.(type) {
	case int8:
		return estimateIntSize(int64(val))
	case int16:
		return estimateIntSize(int64(val))
	case int32:
		return estimateIntSize(int64(val))
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
	case map[string]interface{}:
		return estimateMapSize(val, visited)
	case []interface{}:
		return estimateSliceSize(val, visited)
	default:
		// Separate complex collections to reduce complexity
		if size, handled := estimateCollectionSize(val); handled {
			return size
		}
		return estimateReflect(val, visited)
	}
}

// estimateCollectionSize handles common typed slices/maps to avoid reflection.
func estimateCollectionSize(v interface{}) (int, bool) {
	switch val := v.(type) {
	case []string:
		if len(val) == 0 {
			return 2, true
		}
		size := 1 // [
		for i, s := range val {
			if i > 0 {
				size++
			}
			size += len(s) + 2
		}
		size++ // ]
		return size, true
	case map[string]string:
		if len(val) == 0 {
			return 2, true
		}
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
		return size, true
	case []int:
		if len(val) == 0 {
			return 2, true
		}
		size := 1
		for i, n := range val {
			if i > 0 {
				size++
			}
			size += estimateIntSize(int64(n))
		}
		size++
		return size, true
	case []int64:
		if len(val) == 0 {
			return 2, true
		}
		size := 1
		for i, n := range val {
			if i > 0 {
				size++
			}
			size += estimateIntSize(n)
		}
		size++
		return size, true
	case []float64:
		if len(val) == 0 {
			return 2, true
		}
		size := 1
		var buf [32]byte
		for i, n := range val {
			if i > 0 {
				size++
			}
			b := strconv.AppendFloat(buf[:0], n, 'g', -1, 64)
			size += len(b)
		}
		size++
		return size, true
	}
	return 0, false
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
		// Fallback to fmt.Sprintf only for truly unknown types
		return len(fmt.Sprintf("%v", v)) + 2
	}
}

func estimateStruct(val reflect.Value, visited map[uintptr]bool) int {
	typ := val.Type()
	size := 1 // {
	count := 0

	// ⚡ Bolt: Use cached fields
	fields := getCachedFields(typ)

	for _, fieldMeta := range fields {
		if fieldMeta.IsUnexported {
			continue
		}

		fieldVal := val.Field(fieldMeta.Index)

		if fieldMeta.OmitEmpty && isEmptyValue(fieldVal) {
			continue
		}

		if count > 0 {
			size++ // ,
		}

		size += len(fieldMeta.Name) + 2 // "key"
		size++ // :

		if fieldVal.Kind() == reflect.Ptr || fieldVal.Kind() == reflect.Interface {
			if fieldVal.IsNil() {
				size += 4
			} else {
				size += estimateJSONSizeRecursive(fieldVal.Interface(), visited)
			}
		} else {
			// Optimization: specialized calls for common types to avoid Interface() allocation if possible
			// For now, recursive call handles fast paths for primitives well enough.
			size += estimateJSONSizeRecursive(fieldVal.Interface(), visited)
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
