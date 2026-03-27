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
		case *ast.AssignStmt:
			for _, lhs := range x.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && id.Name == "_" {
					if len(x.Rhs) > 0 {
						if call, ok := x.Rhs[0].(*ast.CallExpr); ok {
                            if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
                                fmt.Printf("Line %d: found '_' assignment for a function call %s\n", fset.Position(x.Pos()).Line, sel.Sel.Name)
                            } else if id, ok := call.Fun.(*ast.Ident); ok {
                                fmt.Printf("Line %d: found '_' assignment for a function call %s\n", fset.Position(x.Pos()).Line, id.Name)
                            }
						}
					}
				}
			}
		case *ast.ExprStmt:
			if call, ok := x.X.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if sel.Sel.Name == "Write" || sel.Sel.Name == "Marshal" || sel.Sel.Name == "Encode" || sel.Sel.Name == "Close" {
						fmt.Printf("Line %d: unassigned call to %s\n", fset.Position(x.Pos()).Line, sel.Sel.Name)
					}
				}
			}
		}
		return true
	})
}
