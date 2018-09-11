package goast

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/flowdev/gflowparser"
	"github.com/flowdev/gflowparser/data"
)

const (
	flowMarker           = "\n\nflow:\n"
	mdStart              = "# Flow Documentation For File: "
	flowStart            = "## Flow: "
	dslMarker            = "    "
	referenceTableHeader = `Components | Data
---------- | -----
`
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
	svg, compTypes, dataTypes, info, err := gflowparser.ConvertFlowDSLToSVG(flow, flowname)
	if err != nil {
		return nil, err
	}
	if info != "" {
		log.Printf("INFO: %s", info)
	}
	if err = ioutil.WriteFile(flowname+".svg", svg, os.FileMode(0666)); err != nil {
		return nil, err
	}
	if _, err = f.WriteString(fmt.Sprintf("![Flow: %s](./%s.svg)\n\n", flowname, flowname)); err != nil {
		return nil, err
	}
	if err = writeReferences(f, compTypes, dataTypes); err != nil {
		return nil, err
	}
	if _, err = f.WriteString(end); err != nil {
		return nil, err
	}

	return f, nil
}

func writeReferences(f *os.File, compTypes []data.Type, dataTypes []data.Type) error {
	dataTypes = filterTypes(dataTypes)
	n := max(len(compTypes), len(dataTypes))
	if n == 0 {
		return nil
	}

	if _, err := f.WriteString(referenceTableHeader); err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		row := bytes.Buffer{}
		if i < len(compTypes) {
			row.WriteString(typeToString(compTypes[i]))
		}
		row.WriteString(" | ")
		if i < len(dataTypes) {
			row.WriteString(typeToString(dataTypes[i]))
		}
		row.WriteRune('\n')
		if _, err := f.Write(row.Bytes()); err != nil {
			return err
		}
	}
	if _, err := f.WriteString("\n"); err != nil {
		return err
	}
	return nil
}
func filterTypes(types []data.Type) []data.Type {
	result := make([]data.Type, 0, len(types))
	for _, t := range types {
		if t.Package != "" {
			result = append(result, t)
			continue
		}
		s := t.LocalType
		if len(s) > 2 && s[:2] == "[]" {
			s = s[2:]
		} else if len(s) > 4 && s[:4] == "map[" {
			continue
		}
		switch s {
		case "bool", "byte", "complex64", "complex128", "float32", "float64",
			"int", "int8", "int16", "int32", "int64",
			"rune", "string", "uint", "uint8", "uint16", "uint32", "uint64",
			"uintptr":
			continue
		default:
			t.LocalType = s
			result = append(result, t)
		}
	}
	return result
}
func typeToString(t data.Type) string {
	if t.Package != "" {
		return t.Package + "." + t.LocalType
	}
	return t.LocalType
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
	start = doc[:i+1]
	i += len(flowMarker)

	buf := bytes.Buffer{}
	for dsl, ok := getDSLLine(doc, &i); ok; dsl, ok = getDSLLine(doc, &i) {
		buf.WriteString(dsl)
	}

	end = doc[i:]
	if end != "" && end[len(end)-1:] != "\n" {
		end += "\n"
	}
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

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
