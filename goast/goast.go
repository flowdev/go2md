package goast

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

const (
	flowMarker = "\n\nflow:\n"
	mdStart    = "# Flow Documentation For File: "
	flowStart  = "## Flow: "
	dslMarker  = "    "
)

// ProcessPackage is processing all the files of one Go package.
func ProcessPackage(pkg *ast.Package) {
	fmt.Println("Processing package:", pkg.Name)
	for name, f := range pkg.Files {
		ProcessFile(f, name)
	}
}

// ProcessFile is processing a Go source file given as *ast.File and writing a
// markdown file with documentation about the Go file.
func ProcessFile(astf *ast.File, goname string) error {
	fmt.Println("Processing file:", goname)
	fmt.Printf("Scope: %#v\n", astf.Scope)
	//for name, astObj := range pkg.Scope.Objects {
	//	fmt.Printf("%s: %T", name, astObj)
	//}
	var osf *os.File
	var err error

	// Print comments for functions and type declarations.
	for _, idecl := range astf.Decls {
		fmt.Printf("decl type: %T\n", idecl)
		switch decl := idecl.(type) {
		case *ast.FuncDecl:
			doc := decl.Doc.Text()
			if strings.Contains(doc, flowMarker) {
				osf, err = addToMDFile(osf, goname, decl.Name.Name, doc)
				if err != nil {
					return err
				}
			}
		case *ast.GenDecl:
			if decl.Tok == token.TYPE && len(decl.Specs) == 1 {
				doc := decl.Doc.Text()
				typ := decl.Specs[0].(*ast.TypeSpec)
				if strings.Contains(doc, flowMarker) {
					osf, err = addToMDFile(osf, goname, typ.Name.Name, doc)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return endMDFile(osf)
}

func startMDFile(goname string) (*os.File, error) {
	mdname := goNameToMD(goname)
	fmt.Println("Writing to:", mdname)

	f, err := os.Create(mdname)
	if err != nil {
		return nil, err
	}

	if _, err = f.WriteString(mdStart + goname + "\n\n"); err != nil {
		return nil, err
	}

	return f, nil
}
func goNameToMD(goname string) string {
	ext := filepath.Ext(goname)
	baseName := goname[:len(goname)-len(ext)]
	return baseName + ".md"
}

func addToMDFile(
	f *os.File,
	goname string,
	flowname string,
	doc string,
) (*os.File, error) {
	var err error
	if f == nil {
		f, err = startMDFile(goname)
		if err != nil {
			return nil, err
		}
	}

	if _, err = f.WriteString(flowStart + flowname + "\n"); err != nil {
		return nil, err
	}
	start, flow, end := ExtractFlowDSL(doc)
	if _, err = f.WriteString(start + "\n"); err != nil {
		return nil, err
	}
	if _, err = f.WriteString(flow + "\n"); err != nil {
		return nil, err
	}
	if _, err = f.WriteString(end); err != nil {
		return nil, err
	}

	return f, nil
}

// ExtractFlowDSL extracts the flow DSL from a documentation comment string.
// The doc string should be given without comment characters.
// This function returns everything before the flow in start,
// the flow DSL itself and everything after it in end.
func ExtractFlowDSL(doc string) (start, flow, end string) {
	i := strings.Index(doc, flowMarker)
	if i < 0 {
		return doc, "", ""
	}
	start = doc[:i]
	fmt.Println("start: >", start, "<")
	i += len(flowMarker)
	fmt.Println("flow start: >", doc[i:i+10], "<")

	buf := bytes.Buffer{}
	for dsl, ok := getDSLLine(doc, &i); ok; dsl, ok = getDSLLine(doc, &i) {
		buf.WriteString(dsl)
	}

	end = doc[i:]
	if end != "" && end[len(end)-1:] != "\n" {
		end += "\n"
	}
	fmt.Println("flow: >", buf.String(), "<")
	fmt.Println("end: >", end, "<")
	return start, buf.String(), end
}
func getDSLLine(doc string, pi *int) (string, bool) {
	if *pi >= len(doc) {
		return "", false
	}
	tail := doc[*pi:]
	n := strings.IndexRune(tail, '\n')
	line := ""
	if n >= 0 {
		n++ // include NL
		line = tail[:n]
	} else {
		n = len(tail)
		line = tail + "\n" // add missing NL
	}

	dslN := len(dslMarker)
	if strings.TrimSpace(line) == "" { // support empty lines
		*pi += n
		return "\n", true
	}
	if n > dslN && line[:dslN] == dslMarker { // real DSL
		*pi += n
		return line[dslN:], true
	}
	return "", false
}

func endMDFile(f *os.File) error {
	if f == nil {
		return nil
	}
	return f.Close()
}
