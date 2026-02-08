// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main checks for missing documentation on exported symbols.
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
		fmt.Println("Usage: go run tools/check_doc.go <directory>")
		os.Exit(1)
	}

	root := os.Args[1]
	fset := token.NewFileSet()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "build" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") || strings.HasSuffix(path, ".pb.gw.go") {
			return nil
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", path, err)
			return nil
		}

		checkFile(f, fset, path)
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking path: %v\n", err)
		os.Exit(1)
	}
}

func checkFile(f *ast.File, fset *token.FileSet, path string) {
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.IsExported() {
				checkDoc(x.Doc, x.Type, x.Name.Name, fset, path, x.Pos())
			}
		case *ast.GenDecl:
			checkGenDecl(x, fset, path)
		}
		return true
	})
}

func checkDoc(doc *ast.CommentGroup, typ *ast.FuncType, name string, fset *token.FileSet, path string, pos token.Pos) {
	if doc == nil {
		fmt.Printf("%s:%d: missing doc for function %s\n", path, fset.Position(pos).Line, name)
		return
	}
	text := doc.Text()

	// Check Summary
	if !strings.Contains(text, "Summary:") {
		fmt.Printf("%s:%d: missing 'Summary:' in doc for function %s\n", path, fset.Position(pos).Line, name)
	}

	// Check Parameters
	if typ.Params != nil && len(typ.Params.List) > 0 {
		if !strings.Contains(text, "Parameters:") {
			fmt.Printf("%s:%d: missing 'Parameters:' in doc for function %s\n", path, fset.Position(pos).Line, name)
		}
	}

	// Check Returns
	if typ.Results != nil && len(typ.Results.List) > 0 {
		if !strings.Contains(text, "Returns:") {
			fmt.Printf("%s:%d: missing 'Returns:' in doc for function %s\n", path, fset.Position(pos).Line, name)
		}

		// Check Errors
		hasError := false
		for _, field := range typ.Results.List {
			if isErrorType(field.Type) {
				hasError = true
				break
			}
		}
		if hasError {
			if !strings.Contains(text, "Errors:") && !strings.Contains(text, "Throws:") {
				fmt.Printf("%s:%d: missing 'Errors:' or 'Throws:' in doc for function %s\n", path, fset.Position(pos).Line, name)
			}
		}
	}
}

func isErrorType(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok && ident.Name == "error" {
		return true
	}
	return false
}

func checkGenDecl(x *ast.GenDecl, fset *token.FileSet, path string) {
	if x.Tok == token.TYPE || x.Tok == token.CONST || x.Tok == token.VAR {
		for _, s := range x.Specs {
			switch ts := s.(type) {
			case *ast.TypeSpec:
				if ts.Name.IsExported() {
					if x.Doc == nil && ts.Doc == nil {
						fmt.Printf("%s:%d: missing doc for type %s\n", path, fset.Position(ts.Pos()).Line, ts.Name.Name)
					}
					// Check methods in interface
					if iface, ok := ts.Type.(*ast.InterfaceType); ok {
						for _, field := range iface.Methods.List {
							if len(field.Names) > 0 && field.Names[0].IsExported() {
								if field.Doc == nil {
									fmt.Printf("%s:%d: missing doc for interface method %s.%s\n", path, fset.Position(field.Pos()).Line, ts.Name.Name, field.Names[0].Name)
								} else {
									// Check doc content for interface methods too
									if funcType, ok := field.Type.(*ast.FuncType); ok {
										checkDoc(field.Doc, funcType, fmt.Sprintf("%s.%s", ts.Name.Name, field.Names[0].Name), fset, path, field.Pos())
									}
								}
							}
						}
					}
				}
			case *ast.ValueSpec:
				for _, name := range ts.Names {
					if name.IsExported() {
						if x.Doc == nil && ts.Doc == nil {
							fmt.Printf("%s:%d: missing doc for var/const %s\n", path, fset.Position(name.Pos()).Line, name.Name)
						}
					}
				}
			}
		}
	}
}
