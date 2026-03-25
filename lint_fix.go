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
        if funcDecl, ok := n.(*ast.FuncDecl); ok && funcDecl.Name.Name == "handleTools" {
            ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
                if exprStmt, ok := n.(*ast.ExprStmt); ok {
                    if call, ok := exprStmt.X.(*ast.CallExpr); ok {
                        if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
                            if sel.Sel.Name == "Write" || sel.Sel.Name == "Marshal" {
                                fmt.Printf("Unchecked return value of %s at %s\n", sel.Sel.Name, fset.Position(exprStmt.Pos()))
                            }
                        }
                    }
                }
                return true
            })
        }
        return true
    })
}
