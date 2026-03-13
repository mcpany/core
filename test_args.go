package main

import (
	"fmt"
)

func main() {
	var a func(t int, b string, c ...string) string
	fmt.Printf("%T\n", a)
}
