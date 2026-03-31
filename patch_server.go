package main

import (
	"io/ioutil"
	"strings"
)

func main() {
	content, err := ioutil.ReadFile("server/pkg/app/api_traces.go")
	if err != nil {
		panic(err)
	}

	str := string(content)
	str = strings.Replace(str, "\"fmt\"", "\"context\"\n\t\"fmt\"", 1)

	err = ioutil.WriteFile("server/pkg/app/api_traces.go", []byte(str), 0644)
	if err != nil {
		panic(err)
	}
}
