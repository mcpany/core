package main

import (
	"bytes"
	"fmt"
	"math"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"github.com/mcpany/core/server/pkg/util"
)

func main() {
	jsonContent := []byte(`{"id": 3000000000}`)
	var inputs map[string]interface{}

	var fastJSON = jsoniter.ConfigCompatibleWithStandardLibrary
	decoder := fastJSON.NewDecoder(bytes.NewReader(jsonContent))
	decoder.UseNumber()
	if err := decoder.Decode(&inputs); err != nil {
		panic(err)
	}

	val := inputs["id"]
	fmt.Printf("Type: %T\n", val)
	fmt.Printf("Value: %v\n", val)
	fmt.Printf("ToString: %s\n", util.ToString(val))

	// Check float64 logic manually
	f := float64(3000000000)
	fmt.Printf("Float64 logic check for %v:\n", f)
	if math.Trunc(f) == f {
		fmt.Println("Is Integer")
		if f >= float64(math.MinInt64) && f < float64(math.MaxInt64) {
			fmt.Printf("In int64 range. Formatted: %s\n", strconv.FormatInt(int64(f), 10))
		} else {
			fmt.Println("Out of int64 range")
		}
	} else {
		fmt.Println("Not Integer")
	}
}
