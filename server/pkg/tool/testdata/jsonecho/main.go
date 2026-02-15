package main

import (
	"encoding/json"
	"io"
	"os"
)

func main() {
	var data map[string]interface{}
	if err := json.NewDecoder(os.Stdin).Decode(&data); err != nil {
		if err != io.EOF {
			os.Exit(1)
		}
	}

	if err := json.NewEncoder(os.Stdout).Encode(data); err != nil {
		os.Exit(1)
	}
}
