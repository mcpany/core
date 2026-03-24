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
        if block, ok := n.(*ast.BlockStmt); ok {
            for _, stmt := range block.List {
                if assign, ok := stmt.(*ast.AssignStmt); ok {
                    for _, lhs := range assign.Lhs {
                        if ident, ok := lhs.(*ast.Ident); ok {
                            if ident.Name == "err" && ident.Obj != nil {
                                // check if err is ignored
                                // fmt.Printf("err assigned at %s\n", fset.Position(assign.Pos()))
                            }
                        }
                    }
                }
            }
        }
		return true
	})
}
