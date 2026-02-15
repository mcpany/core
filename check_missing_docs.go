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
		fmt.Println("Usage: go run check_missing_docs.go <directory>")
		os.Exit(1)
	}

	dir := os.Args[1]
	fset := token.NewFileSet()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			// Skip generated files
			if strings.HasSuffix(path, ".pb.go") || strings.HasSuffix(path, ".pb.gw.go") || strings.Contains(path, "zz_generated") {
				return nil
			}
            // Skip third_party or vendor
            if strings.Contains(path, "vendor/") || strings.Contains(path, "third_party/") {
                return nil
            }

			f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				fmt.Printf("Failed to parse %s: %v\n", path, err)
				return nil
			}

			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncDecl:
					if x.Name.IsExported() {
						if x.Doc == nil {
							fmt.Printf("Missing doc: func %s in %s:%d\n", x.Name.Name, path, fset.Position(x.Pos()).Line)
						}
					}
				case *ast.GenDecl:
					if x.Tok == token.TYPE || x.Tok == token.CONST || x.Tok == token.VAR {
						for _, spec := range x.Specs {
							switch s := spec.(type) {
							case *ast.TypeSpec:
								if s.Name.IsExported() {
									if x.Doc == nil && s.Doc == nil { // Check both decl doc and spec doc
										fmt.Printf("Missing doc: type %s in %s:%d\n", s.Name.Name, path, fset.Position(s.Pos()).Line)
									}
								}
							case *ast.ValueSpec:
								for _, name := range s.Names {
									if name.IsExported() {
										if x.Doc == nil && s.Doc == nil {
											fmt.Printf("Missing doc: var/const %s in %s:%d\n", name.Name, path, fset.Position(name.Pos()).Line)
										}
									}
								}
							}
						}
					}
				case *ast.Field: // Exported fields in structs
                    // This is harder because fields are inside TypeSpec.
                    // But usually GoDoc requires documenting exported fields too.
                    if len(x.Names) > 0 {
                        for _, name := range x.Names {
                            if name.IsExported() && x.Doc == nil {
                                fmt.Printf("Missing doc: field %s in %s:%d\n", name.Name, path, fset.Position(name.Pos()).Line)
                            }
                        }
                    }
				}
				return true
			})
		}
		return nil
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
