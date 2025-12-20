// Package osexitcheck проверяет наличие прямых вызовов os.Exit в функции main пакета main.
// Анализатор сообщает о прямых вызовах os.Exit в main и помогает избежать непредсказуемого завершения процесса.
// Рекомендуется возвращать ошибки из main и централизовать вызов os.Exit.
package osexitcheck

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// Analyzer - это анализатор, который проверяет наличие прямых вызовов os.Exit в функции main пакета main.
var Analyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  "reports direct calls to os.Exit in the main function of package main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg == nil || pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, f := range pass.Files {
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv != nil || fd.Name == nil || fd.Name.Name != "main" || fd.Body == nil {
				continue
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				ce, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				if sel, ok := ce.Fun.(*ast.SelectorExpr); ok {
					if obj := pass.TypesInfo.ObjectOf(sel.Sel); obj != nil {
						if fn, ok := obj.(*types.Func); ok && fn.Pkg() != nil && fn.Pkg().Path() == "os" && fn.Name() == "Exit" {
							pass.Reportf(ce.Pos(), "direct call to os.Exit is forbidden in main")
						}
					}
					return true
				}

				return true
			})
		}
	}

	return nil, nil
}
