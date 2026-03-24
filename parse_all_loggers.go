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
				if sel.Sel.Name == "Error" || sel.Sel.Name == "Info" || sel.Sel.Name == "Warn" || sel.Sel.Name == "Debug" {
					if call2, ok := sel.X.(*ast.CallExpr); ok {
						if sel2, ok := call2.Fun.(*ast.SelectorExpr); ok {
							if sel2.Sel.Name == "GetLogger" {
                                if len(x.Args) % 2 == 0 { // msg + key/val... so 1 + 2n = odd. Even means error
                                    fmt.Printf("Logger call with EVEN args at %s\n", fset.Position(n.Pos()))
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
