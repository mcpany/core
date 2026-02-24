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
				checkFuncDecl(x, fset, path)
			}
		case *ast.GenDecl:
			checkGenDecl(x, fset, path)
		}
		return true
	})
}

func getParamNames(params *ast.FieldList) string {
	var names []string
	if params == nil {
		return ""
	}
	for _, field := range params.List {
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
	}
	return strings.Join(names, ",")
}

func checkFuncDecl(x *ast.FuncDecl, fset *token.FileSet, path string) {
	if x.Doc == nil {
		params := getParamNames(x.Type.Params)
		fmt.Printf("%s:%d: missing doc for function %s (params:%s)\n", path, fset.Position(x.Pos()).Line, x.Name.Name, params)
		return
	}

	text := x.Doc.Text()

	// Check Parameters
	if len(x.Type.Params.List) > 0 {
		if !strings.Contains(text, "Parameters:") {
			params := getParamNames(x.Type.Params)
			fmt.Printf("%s:%d: function %s missing 'Parameters:' section (params:%s)\n", path, fset.Position(x.Pos()).Line, x.Name.Name, params)
		}
	}

	// Check Returns
	if x.Type.Results != nil && len(x.Type.Results.List) > 0 {
		if !strings.Contains(text, "Returns:") {
			// Extract return count/types rough estimation not needed for fix script as we use "result" or "error"
			// But fix script needs to know how many?
			// Actually fix script just appends "Returns: ...".
			// If I can pass return count, that helps.
			count := 0
			for _, r := range x.Type.Results.List {
				if len(r.Names) > 0 {
					count += len(r.Names)
				} else {
					count++
				}
			}
			fmt.Printf("%s:%d: function %s missing 'Returns:' section (count:%d)\n", path, fset.Position(x.Pos()).Line, x.Name.Name, count)
		}
	}

	// Check Errors
	hasError := false
	if x.Type.Results != nil {
		for _, res := range x.Type.Results.List {
			if isErrorType(res.Type) {
				hasError = true
				break
			}
		}
	}
	if hasError {
		if !strings.Contains(text, "Errors:") {
			fmt.Printf("%s:%d: function %s missing 'Errors:' section\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
		}
	}

	// Check Side Effects
	if !strings.Contains(text, "Side Effects:") {
		fmt.Printf("%s:%d: function %s missing 'Side Effects:' section\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
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
									if funcType, ok := field.Type.(*ast.FuncType); ok {
										params := getParamNames(funcType.Params)
										fmt.Printf("%s:%d: missing doc for interface method %s.%s (params:%s)\n", path, fset.Position(field.Pos()).Line, ts.Name.Name, field.Names[0].Name, params)
									} else {
										fmt.Printf("%s:%d: missing doc for interface method %s.%s\n", path, fset.Position(field.Pos()).Line, ts.Name.Name, field.Names[0].Name)
									}
								} else {
									checkInterfaceMethod(field, fset, path, ts.Name.Name)
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

func checkInterfaceMethod(field *ast.Field, fset *token.FileSet, path string, typeName string) {
	text := field.Doc.Text()
	methodName := field.Names[0].Name

	if funcType, ok := field.Type.(*ast.FuncType); ok {
		// Parameters
		if len(funcType.Params.List) > 0 {
			if !strings.Contains(text, "Parameters:") {
				params := getParamNames(funcType.Params)
				fmt.Printf("%s:%d: method %s.%s missing 'Parameters:' section (params:%s)\n", path, fset.Position(field.Pos()).Line, typeName, methodName, params)
			}
		}
		// Returns
		if funcType.Results != nil && len(funcType.Results.List) > 0 {
			if !strings.Contains(text, "Returns:") {
				fmt.Printf("%s:%d: method %s.%s missing 'Returns:' section\n", path, fset.Position(field.Pos()).Line, typeName, methodName)
			}
		}
		// Errors
		hasError := false
		if funcType.Results != nil {
			for _, res := range funcType.Results.List {
				if isErrorType(res.Type) {
					hasError = true
					break
				}
			}
		}
		if hasError {
			if !strings.Contains(text, "Errors:") {
				fmt.Printf("%s:%d: method %s.%s missing 'Errors:' section\n", path, fset.Position(field.Pos()).Line, typeName, methodName)
			}
		}
		// Side Effects
		if !strings.Contains(text, "Side Effects:") {
			fmt.Printf("%s:%d: method %s.%s missing 'Side Effects:' section\n", path, fset.Position(field.Pos()).Line, typeName, methodName)
		}
	}
}
