package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
    b, err := ioutil.ReadFile("server/pkg/app/api.go")
    if err != nil {
        panic(err)
    }
    content := string(b)

    // replacements for the `b, _ := opts.Marshal(obj)` cases that errcheck is mad about
    replacements := [][2]string{
        {
            `b, _ := opts.Marshal(svc)`,
            `b, err := opts.Marshal(svc)
			if err != nil {
				http.Error(w, "Failed to marshal", http.StatusInternalServerError)
				return
			}`,
        },
        {
            `b, _ := opts.Marshal(secret)`,
            `b, err := opts.Marshal(secret)
			if err != nil {
				http.Error(w, "Failed to marshal", http.StatusInternalServerError)
				return
			}`,
        },
        {
            `b, _ := opts.Marshal(exportProfile)`,
            `b, err := opts.Marshal(exportProfile)
			if err != nil {
				http.Error(w, "Failed to marshal", http.StatusInternalServerError)
				return
			}`,
        },
        {
            `b, _ := opts.Marshal(profile)`,
            `b, err := opts.Marshal(profile)
			if err != nil {
				http.Error(w, "Failed to marshal", http.StatusInternalServerError)
				return
			}`,
        },
        {
            `b, _ := opts.Marshal(exportCollection)`,
            `b, err := opts.Marshal(exportCollection)
			if err != nil {
				http.Error(w, "Failed to marshal", http.StatusInternalServerError)
				return
			}`,
        },
        {
            `b, _ := opts.Marshal(collection)`,
            `b, err := opts.Marshal(collection)
			if err != nil {
				http.Error(w, "Failed to marshal", http.StatusInternalServerError)
				return
			}`,
        },
        {
            `b, _ := opts.Marshal(skill)`,
            `b, err := opts.Marshal(skill)
			if err != nil {
				http.Error(w, "Failed to marshal", http.StatusInternalServerError)
				return
			}`,
        },
        {
            `b, _ := opts.Marshal(prompt)`,
            `b, err := opts.Marshal(prompt)
			if err != nil {
				http.Error(w, "Failed to marshal", http.StatusInternalServerError)
				return
			}`,
        },
    }

    for _, repl := range replacements {
        content = strings.ReplaceAll(content, repl[0], repl[1])
    }

    err = ioutil.WriteFile("server/pkg/app/api.go", []byte(content), 0644)
    if err != nil {
        panic(err)
    }
    fmt.Println("Done replacing.")
}
