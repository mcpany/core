package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	err := filepath.Walk("server/pkg", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go") {
			processFile(path)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func processFile(filename string) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing %s: %v\n", filename, err)
		return
	}

	modified := false

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.IsExported() {
				// Check if docstring is missing or incomplete
				// This is a complex logic to implement in Go AST and preserve formatting without breaking comments.
                // We'll just leave this as a thought experiment for now.
			}
		}
		return true
	})
}
