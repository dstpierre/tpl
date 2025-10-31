package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
)

var translationFuncs = map[string]bool{
	"Translate":             true,
	"TranslatePlural":       true,
	"TranslateFormat":       true,
	"TranslateFormatPlural": true,
}

func extractFromCode() ([]string, error) {
	var allKeys []string

	files, err := findAllTemplateFiles("./", "*.go")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		keys, err := extractKeys("tmp.go", string(b))
		if err != nil {
			return nil, err
		}

		allKeys = append(allKeys, keys...)
	}

	return allKeys, nil
}

func extractKeys(name, source string) ([]string, error) {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, name, source, 0)
	if err != nil {
		return nil, err
	}

	keys := []string{}

	ast.Inspect(f, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		var funcName string
		switch fun := callExpr.Fun.(type) {
		case *ast.SelectorExpr:
			// Handles calls like 'tpl.Translate'
			funcName = fun.Sel.Name
		case *ast.Ident:
			// Handles direct calls like 'Translate' (if imported directly)
			funcName = fun.Name
		default:
			return true
		}

		if !translationFuncs[funcName] {
			return true
		}

		if len(callExpr.Args) < 2 {
			return true
		}

		keyArg := callExpr.Args[1]

		if basicLit, isLit := keyArg.(*ast.BasicLit); isLit && basicLit.Kind == token.STRING {
			cleanKey, err := strconv.Unquote(basicLit.Value)
			if err != nil {
				return true
			}
			keys = append(keys, cleanKey)
		} else {
			// This handles cases where the key is a variable or another function call (e.g., tpl.Translate("lang", myKey))
			fmt.Printf("Warning: Key argument for %s is not a simple string literal (Type: %T). Skipping.\n", funcName, keyArg)
		}

		return true
	})

	return keys, nil
}
