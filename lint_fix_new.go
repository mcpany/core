package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "server/pkg/app/api.go", nil, 0)
	if err != nil {
		log.Fatal(err)
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "Error" || sel.Sel.Name == "Info" || sel.Sel.Name == "Warn" {
					if id, ok := sel.X.(*ast.Ident); ok && id.Name == "logger" {
						if len(x.Args) > 1 && len(x.Args)%2 == 0 {
							fmt.Printf("Line %d: slog call has even number of args (odd number of key/value pairs) - %s\n", fset.Position(x.Pos()).Line, id.Name+"."+sel.Sel.Name)
						}
					} else if call, ok := sel.X.(*ast.CallExpr); ok {
						if sel2, ok := call.Fun.(*ast.SelectorExpr); ok {
							if id, ok := sel2.X.(*ast.Ident); ok && id.Name == "logging" && sel2.Sel.Name == "GetLogger" {
								if len(x.Args) > 1 && len(x.Args)%2 == 0 {
									fmt.Printf("Line %d: slog call has even number of args (odd number of key/value pairs) - logging.GetLogger().%s\n", fset.Position(x.Pos()).Line, sel.Sel.Name)
								}
							}
						}
					}
				}
			}
		}
		return true
	})
}
