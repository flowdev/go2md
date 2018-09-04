package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"log"

	"github.com/flowdev/go2md/goast"
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
			goast.ProcessPackage(pkg)
		}
	} else {
		fmt.Println("Parsing file:", flagFile)
		f, err := parser.ParseFile(fset, flagFile, nil, parser.ParseComments)
		if err != nil {
			log.Fatal("Fatal error: Unable to parse the file: " + err.Error())
		}
		goast.ProcessFile(f, flagFile)
	}

}
