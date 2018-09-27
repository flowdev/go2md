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

type sourcePartKind int

const (
	sourcePartUnknown = sourcePartKind(iota)
	sourcePartFlow
	sourcePartFunc
	sourcePartType
)

type sourcePart struct {
	kind   sourcePartKind
	name   string
	doc    string
	start  int
	end    int
	goFile string
	mdFile *mdFile
}

type mdFile struct {
	name   string
	osfile *os.File
}

const (
	markerFlow = "fl:"
	markerFunc = "fu:"
	markerType = "ty:"
)

// ProcessPackage is processing all the files of one Go package.
func ProcessPackage(pkg *ast.Package, fs *token.FileSet) {
	fmt.Println("Processing package:", pkg.Name)
	partMap := make(map[string]*sourcePart)
	flows := make([]*sourcePart, 0, 128)
	fileMap := make(map[string]*mdFile)
	var err error

	for name, f := range pkg.Files {
		if flows, err = findSourceParts(partMap, flows, fileMap, f, fs, name); err != nil {
			log.Fatalf("Unable to find all flows in package: %v", err)
		}
	}
	fmt.Println("Found", len(flows), "flows.")
	for _, f := range flows {
		if err = processFlow(f, partMap); err != nil {
			log.Fatalf("Unable to process all flows in package: %v", err)
		}
	}
	fmt.Println("Processed", len(partMap), "flows.")
	for _, f := range fileMap {
		if err = endMDFile(f.osfile); err != nil {
			log.Printf("Error while ending file: %v", err)
		}
	}
	fmt.Println("Ended", len(fileMap), "files.")
}

func findSourceParts(
	partMap map[string]*sourcePart, flows []*sourcePart, fileMap map[string]*mdFile, astf *ast.File, fs *token.FileSet, goname string,
) ([]*sourcePart, error) {
	var err error
	baseName := goNameToBase(goname)

	// astf.Scope is a dead end just use the Decls instead!
	for _, idecl := range astf.Decls {
		fmt.Printf("decl type: %T\n", idecl)
		switch decl := idecl.(type) {
		case *ast.FuncDecl:
			doc := decl.Doc.Text()
			name := decl.Name.Name
			if strings.Contains(doc, flowMarker) {
				flows, err = addFlow(partMap, flows, fileMap, baseName, decl.Name.Name, doc)
				if err != nil {
					return flows, err
				}
			} else {
				partMap[markerFunc+decl.Name.Name] = &sourcePart{
					kind:   sourcePartFunc,
					name:   name,
					start:  lineFor(decl.Pos(), fs),
					end:    lineFor(decl.End(), fs),
					goFile: goname,
				}
			}
		case *ast.GenDecl:
			if decl.Tok == token.TYPE {
				for _, s := range decl.Specs {
					ts := s.(*ast.TypeSpec)
					name := ts.Name.Name
					partMap[markerType+name] = &sourcePart{
						kind:   sourcePartType,
						name:   name,
						start:  lineFor(ts.Pos(), fs),
						end:    lineFor(ts.End(), fs),
						goFile: goname,
					}
				}
			}
		}
	}
	return flows, nil
}
func goNameToBase(goname string) string {
	ext := filepath.Ext(goname)
	return goname[:len(goname)-len(ext)]
}
func lineFor(p token.Pos, fs *token.FileSet) int {
	if p.IsValid() {
		pos := fs.PositionFor(p, false)
		return pos.Line
	}

	return 0
}

func addFlow(
	flowMap map[string]*sourcePart, flows []*sourcePart, fileMap map[string]*mdFile, fileBaseName string,
	fname, doc string,
) ([]*sourcePart, error) {
	if i := strings.Index(fname, "_"); i >= 0 {
		fname = fname[:i] // cut off the port name
	}
	file := fileMap[fileBaseName]
	if file == nil {
		osfile, err := startMDFile(fileBaseName)
		if err != nil {
			return flows, err
		}
		file = &mdFile{name: fileBaseName, osfile: osfile}
		fileMap[fileBaseName] = file
	}
	flow := &sourcePart{kind: sourcePartFlow, name: fname, doc: doc, mdFile: file}
	flows = append(flows, flow)
	flowMap[fname] = flow
	fmt.Println("Found", len(flows), "flows.")
	return flows, nil
}

func startMDFile(fileBaseName string) (*os.File, error) {
	mdname := fileBaseName + ".md"
	fmt.Println("Opening file:", mdname)

	f, err := os.Create(mdname)
	if err != nil {
		return nil, err
	}

	if _, err = f.WriteString(mdStart + fileBaseName + ".go\n\n"); err != nil {
		return nil, err
	}

	return f, nil
}

func processFlow(f *sourcePart, flowMap map[string]*sourcePart) error {
	fmt.Println("Processing flow:", f.name)
	err := addToMDFile(f, flowMap)
	return err
}

func addToMDFile(f *sourcePart, flowMap map[string]*sourcePart) error {
	if _, err := f.mdFile.osfile.WriteString(flowStart + f.name + "\n"); err != nil {
		return err
	}
	start, flow, end := ExtractFlowDSL(f.doc)
	if _, err := f.mdFile.osfile.WriteString(start + "\n"); err != nil {
		return err
	}
	svg, compTypes, dataTypes, info, err := gflowparser.ConvertFlowDSLToSVG(flow, f.name)
	if err != nil {
		return err
	}
	if info != "" {
		log.Printf("INFO: %s", info)
	}
	if err = ioutil.WriteFile(f.name+".svg", svg, os.FileMode(0666)); err != nil {
		return err
	}
	if _, err = f.mdFile.osfile.WriteString(fmt.Sprintf("![Flow: %s](./%s.svg)\n\n", f.name, f.name)); err != nil {
		return err
	}
	if err = writeReferences(f, compTypes, dataTypes, flowMap); err != nil {
		return err
	}
	if _, err = f.mdFile.osfile.WriteString(end); err != nil {
		return err
	}

	return nil
}

func writeReferences(
	f *sourcePart, compTypes []data.Type,
	dataTypes []data.Type,
	flowMap map[string]*sourcePart) error {
	dataTypes = filterTypes(dataTypes)
	n := max(len(compTypes), len(dataTypes))
	if n == 0 {
		return nil
	}

	if _, err := f.mdFile.osfile.WriteString(referenceTableHeader); err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		row := bytes.Buffer{}
		if i < len(compTypes) {
			addComponentToRow(&row, compTypes[i], flowMap)
		}
		row.WriteString(" | ")
		if i < len(dataTypes) {
			row.WriteString(typeToString(dataTypes[i]))
		}
		row.WriteRune('\n')
		if _, err := f.mdFile.osfile.Write(row.Bytes()); err != nil {
			return err
		}
	}
	if _, err := f.mdFile.osfile.WriteString("\n"); err != nil {
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
func addComponentToRow(row *bytes.Buffer, comp data.Type, flowMap map[string]*sourcePart) {
	cNam := typeToString(comp)
	flow := flowMap[cNam]
	if flow != nil {
		// [link to Google!](http://google.com)
		row.WriteString(
			"[" + cNam + "](" +
				"./" + flow.mdFile.name + ".md#flow-" +
				strings.ToLower(flow.name) +
				")")
	} else {
		row.WriteString(cNam)
	}
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
