package main

import (
	"fmt"
	"google.golang.org/protobuf/proto"
)

func main() {
	s := proto.String("hello")
	fmt.Println(*s)
}
