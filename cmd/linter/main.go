package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

var ErrCheckAnalyzer = &analysis.Analyzer{
	Name: "errcheck",
	Doc:  "check for unchecked errors",
	Run:  run,
}

func main() {
	singlechecker.Main(ErrCheckAnalyzer)
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		findPanicOrOsExitOrLogFatal(pass, file)
	}

	return nil, nil
}

func findPanicOrOsExitOrLogFatal(pass *analysis.Pass, file ast.Node) {
	var currentRunningFunction ast.FuncDecl
	ast.Inspect(file, func(node ast.Node) bool {
		function, ok := node.(*ast.FuncDecl)
		if ok {
			currentRunningFunction = *function
		}

		functionCallExpression, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		// panic check
		functionExpression, ok := functionCallExpression.Fun.(*ast.Ident)
		if ok && functionExpression.Name == "panic" {
			pass.Reportf(functionExpression.NamePos, "panic function call!")
			return true
		}

		functionSelectorExpression, ok := functionCallExpression.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		packageIdentifier, ok := functionSelectorExpression.X.(*ast.Ident)
		if !ok || functionSelectorExpression.Sel == nil {
			return true
		}

		// os.Exit() and log.Fatal() calls check
		if (packageIdentifier.Name == "os" && functionSelectorExpression.Sel.Name == "Exit") ||
			(packageIdentifier.Name == "log" && functionSelectorExpression.Sel.Name == "Fatal") {
			if !isMainFunction(pass, &currentRunningFunction) {
				pass.Reportf(functionSelectorExpression.Sel.NamePos,
					"log.Fatal or os.Exit calls outside of main function are prohibited!")
			}
		}

		return true
	})
}

func isMainFunction(pass *analysis.Pass, currentRunningFunction *ast.FuncDecl) bool {
	packageName := pass.Pkg.Name()
	functionName := currentRunningFunction.Name.Name

	if packageName == "main" && functionName == "main" {
		return true
	}

	return false
}
