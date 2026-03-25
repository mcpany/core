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
            for _, lhs := range assign.Lhs {
                if ident, ok := lhs.(*ast.Ident); ok {
                    if ident.Name == "err" {
                        if call, ok := assign.Rhs[0].(*ast.CallExpr); ok {
                            if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
                                if sel.Sel.Name == "Marshal" {
                                    // check if it's assigned with something else being _
                                    if len(assign.Lhs) > 1 {
                                        if id2, ok := assign.Lhs[0].(*ast.Ident); ok && id2.Name == "_" {
                                            fmt.Printf("Marshal returning err but discarding result at %s\n", fset.Position(assign.Pos()))
                                        }
                                    }
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
