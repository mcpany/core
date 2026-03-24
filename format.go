package main

import (
	"fmt"
	"go/format"
	"os"
)

func main() {
	src, err := os.ReadFile("server/pkg/app/api.go")
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := format.Source(src)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = os.WriteFile("server/pkg/app/api.go", res, 0644)
	if err != nil {
		fmt.Println(err)
	}
}
