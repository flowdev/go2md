package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

type (
	t1 int
	t2 string
	t3 float64
)

func main() {
	fname := "./sample.go"
	if len(os.Args) > 1 {
		fname = os.Args[1]
	}
	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	fmt.Printf("Parsing file: %s\n", fname)
	f, err := parser.ParseFile(fset, fname, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing file:\n")
		panic(err)
	}

	// Create an ast.CommentMap from the ast.File's comments.
	// This helps keeping the association between comments
	// and AST nodes.
	cmap := ast.NewCommentMap(fset, f, f.Comments)

	// Remove the first variable declaration from the list of declarations.
	for _, idecl := range f.Decls {
		fmt.Printf("decl type: %T\n", idecl)
		switch decl := idecl.(type) {
		case *ast.FuncDecl:
			fmt.Printf("Function: %s\n", decl.Name)
			fmt.Print(decl.Doc.Text())
			//printComments(cmap[decl])
			fmt.Println()
		case *ast.GenDecl:
			if decl.Tok == token.TYPE {
				fmt.Println("Type (comment):")
				fmt.Print(decl.Doc.Text())
				//printComments(cmap[decl])
				for _, spec := range decl.Specs {
					typ := spec.(*ast.TypeSpec)
					fmt.Printf("Type: %s\n", typ.Name)
					printComments(cmap[typ])
					fmt.Println()
				}
			}
		}
	}

	// Use the comment map to filter comments that don't belong anymore
	// (the comments associated with the variable declaration), and create
	// the new comments list.
	//f.Comments = cmap.Filter(f).Comments()
	/*
		// Print the modified AST.
		var buf bytes.Buffer
		if err := format.Node(&buf, fset, f); err != nil {
			panic(err)
		}
		fmt.Printf("%s", buf.Bytes())
	*/
}

func printComments(comments []*ast.CommentGroup) {
	for _, c := range comments {
		fmt.Print(c.Text())
	}
}
