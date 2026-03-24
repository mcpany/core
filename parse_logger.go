package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "server/pkg/app/api.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "Error" || sel.Sel.Name == "Info" || sel.Sel.Name == "Warn" {
					if call2, ok := sel.X.(*ast.CallExpr); ok {
						if sel2, ok := call2.Fun.(*ast.SelectorExpr); ok {
							if sel2.Sel.Name == "GetLogger" {
								if len(x.Args) % 2 == 0 {
									fmt.Printf("Logger call with EVEN args (msg + k/v) at %s: ", fset.Position(n.Pos()))
                                    for _, arg := range x.Args {
                                        if lit, ok := arg.(*ast.BasicLit); ok {
                                            fmt.Printf("%s, ", lit.Value)
                                        } else if ident, ok := arg.(*ast.Ident); ok {
                                            fmt.Printf("%s, ", ident.Name)
                                        } else {
                                            fmt.Printf("<expr>, ")
                                        }
                                    }
                                    fmt.Println()
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
