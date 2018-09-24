package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"log"

	"github.com/flowdev/go2md/goast"
)

func main() {
	fset := token.NewFileSet() // needed for any kind of parsing
	fmt.Println("Parsing the whole package.")
	pkgs, err := parser.ParseDir(fset, ".", nil, parser.ParseComments)
	if err != nil {
		log.Fatal("Fatal error: Unable to parse the package: " + err.Error())
	}
	for _, pkg := range pkgs { // iterate over subpackages (e.g.: xxx and xxx_test)
		goast.ProcessPackage(pkg)
	}
}
