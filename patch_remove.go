package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	b, err := ioutil.ReadFile("server/pkg/tool/types.go")
	if err != nil {
		panic(err)
	}

	content := string(b)

	// Remove checkUnquotedKeywords_old
	idx := strings.Index(content, "func checkUnquotedKeywords_old(val string, keywords []string) error {")
	if idx != -1 {
	    // Find the end of this function
	    endStr := "\n\t}\n\treturn nil\n}\n\nfunc checkKeyword"
	    endIdx := strings.Index(content[idx:], endStr)
	    if endIdx != -1 {
	        content = content[:idx] + content[idx+endIdx+len("\n\t}\n\treturn nil\n}\n\n"):]
	    }
	}

	ioutil.WriteFile("server/pkg/tool/types.go", []byte(content), 0644)
	fmt.Println("Done removing old function")
}
