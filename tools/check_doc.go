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

		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				if x.Name.IsExported() {
					if x.Doc == nil {
						fmt.Printf("%s:%d: missing doc for function %s\n", path, fset.Position(x.Pos()).Line, x.Name.Name)
					}
				}
			case *ast.GenDecl:
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
											}
										}
									}
								}
								// Check fields in struct? The prompt says "every single public function, method, and class".
								// "class" usually maps to struct/interface in Go.
								// It doesn't explicitly say "every exported field of a struct", but it's good practice.
								// I'll stick to types and functions first.
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
			return true
		})

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking path: %v\n", err)
		os.Exit(1)
	}
}
