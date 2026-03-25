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
        if assign, ok := n.(*ast.AssignStmt); ok {
            if len(assign.Lhs) == 1 {
                for _, lhs := range assign.Lhs {
                    if ident, ok := lhs.(*ast.Ident); ok {
                        if ident.Name == "_" {
                            if call, ok := assign.Rhs[0].(*ast.CallExpr); ok {
                                if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
                                    fmt.Printf("Ignoring entire return value of %s at %s\n", sel.Sel.Name, fset.Position(assign.Pos()))
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
