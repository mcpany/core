package main

import (
	"fmt"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func main() {
	user := configv1.User_builder{
		Id:          proto.String("test"),
		Preferences: map[string]string{"foo": "bar"},
	}.Build()
	fmt.Println(user.GetId())
}
