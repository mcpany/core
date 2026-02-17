// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main checks for missing or non-compliant documentation on exported symbols.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

var strictMode = flag.Bool("strict", false, "Enable strict documentation structure checks")

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("Usage: go run tools/check_doc.go [-strict] <directory>")
		os.Exit(1)
	}

	root := args[0]
	fset := token.NewFileSet()
	hasErrors := false

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "build" || info.Name() == "third_party" || info.Name() == "tests" || info.Name() == "tools" || info.Name() == "examples" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") || strings.HasSuffix(path, ".pb.gw.go") || strings.HasSuffix(path, "zz_generated.deepcopy.go") {
			return nil
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", path, err)
			return nil
		}

		if !checkFile(f, fset, path) {
			hasErrors = true
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking path: %v\n", err)
		os.Exit(1)
	}

	if hasErrors {
		os.Exit(1)
	}
}

func checkFile(f *ast.File, fset *token.FileSet, path string) bool {
	valid := true
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.IsExported() {
				if x.Doc == nil {
					fmt.Printf("%s:%d: missing doc for function %s\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
					valid = false
				} else if *strictMode {
					if !checkFuncDoc(x, path, fset) {
						valid = false
					}
				}
			}
		case *ast.GenDecl:
			if !checkGenDecl(x, fset, path) {
				valid = false
			}
		}
		return true
	})
	return valid
}

func checkFuncDoc(fn *ast.FuncDecl, path string, fset *token.FileSet) bool {
	text := fn.Doc.Text()
	missing := []string{}

	if !strings.Contains(text, "Summary:") {
		missing = append(missing, "Summary")
	}

	if fn.Type.Params != nil && len(fn.Type.Params.List) > 0 {
		if !strings.Contains(text, "Parameters:") {
			missing = append(missing, "Parameters")
		}
	}

	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		if !strings.Contains(text, "Returns:") {
			missing = append(missing, "Returns")
		}
	}

	// Check for error return
	returnsError := false
	if fn.Type.Results != nil {
		for _, r := range fn.Type.Results.List {
			if ident, ok := r.Type.(*ast.Ident); ok && ident.Name == "error" {
				returnsError = true
				break
			}
		}
	}

	if returnsError {
		if !strings.Contains(text, "Throws:") && !strings.Contains(text, "Errors:") {
			missing = append(missing, "Throws/Errors")
		}
	}

	if len(missing) > 0 {
		fmt.Printf("%s:%d: incomplete doc for function %s (missing: %s)\n", path, fset.Position(fn.Pos()).Line, fn.Name.Name, strings.Join(missing, ", "))
		return false
	}
	return true
}

func checkGenDecl(x *ast.GenDecl, fset *token.FileSet, path string) bool {
	valid := true
	if x.Tok == token.TYPE || x.Tok == token.CONST || x.Tok == token.VAR {
		for _, s := range x.Specs {
			switch ts := s.(type) {
			case *ast.TypeSpec:
				if ts.Name.IsExported() {
					if x.Doc == nil && ts.Doc == nil {
						fmt.Printf("%s:%d: missing doc for type %s\n", path, fset.Position(ts.Pos()).Line, ts.Name.Name)
						valid = false
					} else if *strictMode {
						doc := x.Doc
						if ts.Doc != nil {
							doc = ts.Doc
						}
						if !strings.Contains(doc.Text(), "Summary:") {
							fmt.Printf("%s:%d: incomplete doc for type %s (missing: Summary)\n", path, fset.Position(ts.Pos()).Line, ts.Name.Name)
							valid = false
						}
					}
					// Check methods in interface
					if iface, ok := ts.Type.(*ast.InterfaceType); ok {
						for _, field := range iface.Methods.List {
							if len(field.Names) > 0 && field.Names[0].IsExported() {
								if field.Doc == nil {
									fmt.Printf("%s:%d: missing doc for interface method %s.%s\n", path, fset.Position(field.Pos()).Line, ts.Name.Name, field.Names[0].Name)
									valid = false
								} else if *strictMode {
									text := field.Doc.Text()
									missing := []string{}
									if !strings.Contains(text, "Summary:") {
										missing = append(missing, "Summary")
									}
									if fType, ok := field.Type.(*ast.FuncType); ok {
										if fType.Params != nil && len(fType.Params.List) > 0 && !strings.Contains(text, "Parameters:") {
											missing = append(missing, "Parameters")
										}
										if fType.Results != nil && len(fType.Results.List) > 0 && !strings.Contains(text, "Returns:") {
											missing = append(missing, "Returns")
										}
									}
									if len(missing) > 0 {
										fmt.Printf("%s:%d: incomplete doc for interface method %s.%s (missing: %s)\n", path, fset.Position(field.Pos()).Line, ts.Name.Name, field.Names[0].Name, strings.Join(missing, ", "))
										valid = false
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
							valid = false
						} else if *strictMode {
							doc := x.Doc
							if ts.Doc != nil {
								doc = ts.Doc
							}
							if !strings.Contains(doc.Text(), "Summary:") {
								fmt.Printf("%s:%d: incomplete doc for var/const %s (missing: Summary)\n", path, fset.Position(name.Pos()).Line, name.Name)
								valid = false
							}
						}
					}
				}
			}
		}
	}
	return valid
}
