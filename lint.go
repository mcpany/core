package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "server/pkg/app/api_test.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, decl := range f.Decls {
		if fn, isFn := decl.(*ast.FuncDecl); isFn {
			if fn.Doc == nil {
				fmt.Printf("Missing doc: %s\n", fn.Name.Name)
			}
		}
	}
}
