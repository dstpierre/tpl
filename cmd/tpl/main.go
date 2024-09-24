package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"os"
	"text/template/parse"
)

func printTypes(f *ast.File) {
	ast.Walk(visitor{}, f)
}

type visitor struct{}

func (v visitor) Visit(node ast.Node) ast.Visitor {
	switch t := node.(type) {
	case *ast.TypeSpec:
		if spec, ok := t.Type.(*ast.StructType); ok {
			fmt.Printf("Type: %s\n", t.Name.Name)
			for _, field := range spec.Fields.List {
				fmt.Printf("  Field: %s\n", field.Names[0].Name)
				/*if len(field.Type.(*ast.StarExpr).Expr.(*ast.Ident).Name) > 0 {
					fmt.Printf("    Type: %s\n", field.Type.(*ast.StarExpr).Expr.(*ast.Ident).Name)
				} else {
					fmt.Printf("    Type: %s\n", field.Type.(*ast.Ident).Name)
				}*/
			}
		}
	}
	return v
}

func main() {
	pwd, err := os.Executable()
	if err != nil {
		fmt.Println("cannot get current directory: ", err)
		os.Exit(1)
	}

	var dir string
	var tsrc string

	flag.StringVar(&dir, "src", pwd, "directory of the root source Go pacakge")
	flag.StringVar(&tsrc, "t", "", "path for the template to check")
	flag.Parse()

	tree, err := getTree(tsrc)
	if err != nil {
		fmt.Println("error parsing template: ", err)
		os.Exit(1)
	}

	fset := token.NewFileSet()
	pkg, err := parser.ParseDir(fset, dir, nil, 0)
	if err != nil {
		fmt.Println("error parsing source directory: ", err)
		os.Exit(1)
	}

	for _, v := range pkg {
		for _, f := range v.Files {
			printTypes(f)
		}
	}

	fmt.Println("======")
	fmt.Println(tree)
	fmt.Println("======")
}

func getTree(filename string) (*parse.Tree, error) {
	t, err := template.ParseFiles(filename)
	if err != nil {
		return nil, err
	}

	return t.Tree, nil
}
