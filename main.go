package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

func main() {
	var flagFile string
	var flagPackage bool
	const (
		usageFile    = "the file to search for flow documentation"
		usagePackage = "search the whole package for flow documentation"
	)
	flag.StringVar(&flagFile, "file", "", usageFile)
	flag.StringVar(&flagFile, "f", "", usageFile+" (shorthand)")
	flag.BoolVar(&flagPackage, "package", false, usagePackage)
	flag.BoolVar(&flagPackage, "p", false, usagePackage+" (shorthand)")
	flag.Parse()

	if flagFile != "" && flagPackage {
		log.Fatal("Fatal error: Unable to process a whole package and a single file at the same time.")
	}
	if flagFile == "" && !flagPackage {
		log.Fatal("Fatal error: Neither a whole package nor a single file is given.")
	}

	fset := token.NewFileSet() // needed for any kind of parsing
	if flagPackage {
		fmt.Println("Parsing the whole package.")
		pkgs, err := parser.ParseDir(fset, ".", nil, parser.ParseComments)
		if err != nil {
			log.Fatal("Fatal error: Unable to parse the package: " + err.Error())
		}
		for _, pkg := range pkgs {
			processPackage(pkg)
		}
	} else {
		fmt.Println("Parsing file:", flagFile)
		f, err := parser.ParseFile(fset, flagFile, nil, parser.ParseComments)
		if err != nil {
			log.Fatal("Fatal error: Unable to parse the file: " + err.Error())
		}
		processFile(f, flagFile)
	}

}

func processPackage(pkg *ast.Package) {
	for name, f := range pkg.Files {
		processFile(f, name)
	}
}
func processFile(f *ast.File, fname string) {
	fmt.Println("Processing file:", fname)

	// Print comments for functions and type declarations.
	for _, idecl := range f.Decls {
		fmt.Printf("decl type: %T\n", idecl)
		switch decl := idecl.(type) {
		case *ast.FuncDecl:
			fmt.Printf("Function: %s\n", decl.Name)
			fmt.Print(decl.Doc.Text())
			fmt.Println()
		case *ast.GenDecl:
			if decl.Tok == token.TYPE {
				fmt.Println("Type (comment):")
				fmt.Print(decl.Doc.Text())
				for _, spec := range decl.Specs {
					typ := spec.(*ast.TypeSpec)
					fmt.Printf("Type: %s\n", typ.Name)
					fmt.Print(typ.Doc.Text())
					fmt.Println()
				}
			}
		}
	}
}
