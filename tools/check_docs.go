// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: check_docs <directory>")
		os.Exit(1)
	}

	rootDir := os.Args[1]
	fset := token.NewFileSet()
	hasError := false

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Skip generated files
		if strings.Contains(path, ".pb.go") || strings.Contains(path, ".pb.gw.go") {
			return nil
		}

		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", path, err)
			return nil
		}

		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				if x.Name.IsExported() {
					if x.Doc == nil || len(x.Doc.List) == 0 {
						fmt.Printf("%s:%d: function %s is missing documentation\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
						hasError = true
					}
				}
			case *ast.GenDecl:
				if x.Tok == token.TYPE || x.Tok == token.CONST || x.Tok == token.VAR {
					if x.Doc != nil {
						return true
					}
					for _, spec := range x.Specs {
						switch s := spec.(type) {
						case *ast.TypeSpec:
							if s.Name.IsExported() {
								if s.Doc == nil {
									fmt.Printf("%s:%d: type %s is missing documentation\n", path, fset.Position(s.Pos()).Line, s.Name.Name)
									hasError = true
								}
							}
						case *ast.ValueSpec:
							for _, name := range s.Names {
								if name.IsExported() {
									if s.Doc == nil {
										fmt.Printf("%s:%d: value %s is missing documentation\n", path, fset.Position(s.Pos()).Line, name.Name)
										hasError = true
									}
								}
							}
						}
					}
				}
			}
			return true
		})

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	if hasError {
		os.Exit(1)
	}
}
