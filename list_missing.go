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
	root := "server"
	fset := token.NewFileSet()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
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
		if err != nil { return nil }

		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				if x.Name.IsExported() {
					if !hasGoldStandard(x.Doc) {
						fmt.Printf("%s: func %s\n", path, x.Name.Name)
					}
				}
			case *ast.GenDecl:
				if x.Tok == token.TYPE || x.Tok == token.CONST || x.Tok == token.VAR {
					for _, s := range x.Specs {
						switch ts := s.(type) {
						case *ast.TypeSpec:
							if ts.Name.IsExported() {
								if !hasGoldStandard(x.Doc) && !hasGoldStandard(ts.Doc) {
									fmt.Printf("%s: type %s\n", path, ts.Name.Name)
								}
								if iface, ok := ts.Type.(*ast.InterfaceType); ok {
									if iface.Methods != nil {
										for _, field := range iface.Methods.List {
											if len(field.Names) > 0 && field.Names[0].IsExported() {
												if !hasGoldStandard(field.Doc) {
													fmt.Printf("%s: interface method %s.%s\n", path, ts.Name.Name, field.Names[0].Name)
												}
											}
										}
									}
								}
							}
						case *ast.ValueSpec:
							for _, name := range ts.Names {
								if name.IsExported() {
									if !hasGoldStandard(x.Doc) && !hasGoldStandard(ts.Doc) {
										fmt.Printf("%s: var/const %s\n", path, name.Name)
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

	if err != nil { fmt.Println(err); os.Exit(1) }
}

func hasGoldStandard(doc *ast.CommentGroup) bool {
	if doc == nil { return false }
	text := doc.Text()
	return strings.Contains(text, "Summary:")
}
