package goast

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
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
	goTestFileName = `_test.go`
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

//
// packageDict is a simple dictionary of all known packages/paths and
// their source parts.
//

type goPackage struct {
	path    string
	partMap map[string]*sourcePart
}

type packageDict struct {
	packs    map[string]*goPackage
	srcRoots []string
}

// NewPackageDict creates a new dictionary for packages
func NewPackageDict(srcRoots []string) *packageDict {
	return &packageDict{
		packs:    make(map[string]*goPackage),
		srcRoots: srcRoots,
	}
}

func (pd *packageDict) addPackage(path string, partMap map[string]*sourcePart) {
	pd.packs[path] = &goPackage{path: path, partMap: partMap}
}
func (pd *packageDict) getPartFor(path string, markedName string) *sourcePart {
	goPack := pd.packs[path]
	if goPack == nil {
		return nil
	}
	return goPack.partMap[markedName]
}

//
// fileImps
//

type fileImps struct {
	imps     map[string]string // maps local package name (without '.') to import path
	packDict *packageDict
	fset     *token.FileSet
}

func newFileImps(
	astImps []*ast.ImportSpec,
	packDict *packageDict,
	fset *token.FileSet,
) *fileImps {
	imps := make(map[string]string)
	for _, astImp := range astImps {
		key := ""
		val := astImp.Path.Value
		if astImp.Name == nil {
			key = path.Base(val)
		} else {
			key = strings.TrimRight(astImp.Name.Name, ".")
		}
		if key != "_" && key != "" && key != "." && key != "/" { // ignore funny imports
			imps[key] = val
		}
	}
	return &fileImps{imps: imps, packDict: packDict}
}
func (fi *fileImps) getPartFor(pack string, markedName string) *sourcePart {
	path := fi.imps[pack]
	if path == "" {
		return nil
	}
	part := fi.packDict.getPartFor(path, markedName)
	if part != nil {
		return part
	}
	partMap := fi.findPartsForPath(path)
	if partMap != nil {
		fi.packDict.addPackage(path, partMap)
	}
	return partMap[markedName]
}
func (fi *fileImps) findPartsForPath(path string) map[string]*sourcePart {
	dir := path
	if dir[0] != '.' {
		for _, baseDir := range fi.packDict.srcRoots {
			absDir := filepath.Join(baseDir, dir)
			finfo, err := os.Stat(absDir)
			if err != nil {
				continue
			}
			if finfo.IsDir() {
				dir = absDir
				break
			}
		}
	}
	pkgs, err := parser.ParseDir(fi.fset, dir, excludeTests, parser.ParseComments)
	if err != nil {
		log.Printf("ERROR: Unable to parse additional directory '%s': %v", dir, err)
		return nil
	}
	partMap := make(map[string]*sourcePart)
	flows := make([]*sourcePart, 0, 128)
	for _, pkg := range pkgs { // iterate over subpackages (e.g.: xxx and xxx_test)
		if len(pkg.Name) >= 5 && pkg.Name[len(pkg.Name)-5:] == "_test" {
			continue
		}
		for name, astf := range pkg.Files {
			if flows, err = findSourceParts(
				partMap, flows,
				astf,
				name, fi.fset,
			); err != nil {
				log.Printf(
					"ERROR: Unable to find all source parts in directory '%s': %v",
					dir, err)
			}
		}
	}
	return partMap
}

//
// Parse and process the main directory/package.
//

func excludeTests(fi os.FileInfo) bool {
	nam := strings.ToLower(fi.Name())
	return !fi.IsDir() &&
		(len(nam) < len(goTestFileName) ||
			nam[len(nam)-len(goTestFileName):] != goTestFileName)
}

// ProcessDir processes the whole given directory
func ProcessDir(dir string, packDict *packageDict) error {
	fset := token.NewFileSet() // needed for any kind of parsing
	fmt.Println("Parsing the whole directory:", dir)
	pkgs, err := parser.ParseDir(fset, dir, excludeTests, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("unable to parse the directory '%s': %v", dir, err)
	}
	for _, pkg := range pkgs { // iterate over subpackages (e.g.: xxx and xxx_test)
		if len(pkg.Name) >= 5 && pkg.Name[len(pkg.Name)-5:] == "_test" {
			continue
		}
		processPackage(pkg, fset)
	}
	return nil
}

// processPackage is processing all the files of one Go package.
func processPackage(pkg *ast.Package, fset *token.FileSet) error {
	fmt.Println("processing package:", pkg.Name)
	partMap := make(map[string]*sourcePart)
	flows := make([]*sourcePart, 0, 128)
	fileMap := make(map[string]*mdFile)
	var err error

	for name, astf := range pkg.Files {
		if flows, err = findSourceParts(
			partMap, flows,
			astf,
			name, fset,
		); err != nil {
			return fmt.Errorf(
				"unable to find all flows in package (%s): %v", pkg.Name, err)
		}
	}
	fmt.Println("Found", len(flows), "flows.")
	for _, f := range flows {
		if err = startFlowFile(f, fileMap); err != nil {
			return fmt.Errorf(
				"unable to start all Markdown files in package (%s): %v",
				pkg.Name, err)
		}
		if err = addToMDFile(f, partMap); err != nil {
			return fmt.Errorf(
				"unable to process all flows in package (%s): %v", pkg.Name, err)
		}
	}
	fmt.Println("processed flows with ", len(partMap), "souce parts.")
	for _, f := range fileMap {
		if err = endMDFile(f); err != nil {
			log.Printf("Error while ending file: %v", err)
		}
	}
	fmt.Println("Ended", len(fileMap), "files.")
	return nil
}

//
// Handle source file
//

func findSourceParts(
	partMap map[string]*sourcePart, flows []*sourcePart,
	astf *ast.File,
	goname string, fset *token.FileSet,
) ([]*sourcePart, error) {
	baseName := goNameToBase(goname)

	for _, idecl := range astf.Decls {
		switch decl := idecl.(type) {
		case *ast.FuncDecl:
			doc := decl.Doc.Text()
			name := decl.Name.Name
			if strings.Contains(doc, flowMarker) {
				if i := strings.Index(name, "_"); i >= 0 {
					name = name[:i] // cut off the port name
				}
				flow := &sourcePart{
					kind:   sourcePartFlow,
					name:   name,
					doc:    doc,
					start:  lineFor(decl.Pos(), fset),
					end:    lineFor(decl.End(), fset),
					mdFile: &mdFile{name: baseName},
				}
				partMap[markerFlow+name] = flow
				flows = append(flows, flow)
			} else {
				partMap[markerFunc+decl.Name.Name] = &sourcePart{
					kind:   sourcePartFunc,
					name:   name,
					start:  lineFor(decl.Pos(), fset),
					end:    lineFor(decl.End(), fset),
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
						start:  lineFor(ts.Pos(), fset),
						end:    lineFor(ts.End(), fset),
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
func lineFor(p token.Pos, fset *token.FileSet) int {
	if p.IsValid() {
		pos := fset.PositionFor(p, false)
		return pos.Line
	}

	return 0
}

//
// Write to Markdown file
//

func startFlowFile(flow *sourcePart, fileMap map[string]*mdFile) error {
	file := fileMap[flow.mdFile.name]
	if file == nil {
		osfile, err := startMDFile(flow.mdFile.name)
		if err != nil {
			return err
		}
		file = &mdFile{name: flow.mdFile.name, osfile: osfile}
		fileMap[flow.mdFile.name] = file
	}
	flow.mdFile = file
	return nil
}

func startMDFile(fileBaseName string) (*os.File, error) {
	mdname := fileBaseName + ".md"

	f, err := os.Create(mdname)
	if err != nil {
		return nil, err
	}

	if _, err = f.WriteString(mdStart + fileBaseName + ".go\n\n"); err != nil {
		return nil, err
	}

	return f, nil
}

func addToMDFile(f *sourcePart, partMap map[string]*sourcePart) error {
	fmt.Println("processing flow:", f.name)
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
	if err = writeReferences(f, compTypes, dataTypes, partMap); err != nil {
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
	partMap map[string]*sourcePart,
) error {
	dataTypes = filterTypes(dataTypes)
	dataTypes = sortTypes(dataTypes)
	compTypes = sortTypes(compTypes)

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
			addComponentToRow(&row, compTypes[i], partMap)
		}
		row.WriteString(" | ")
		if i < len(dataTypes) {
			addTypeToRow(&row, dataTypes[i], partMap)
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
func sortTypes(types []data.Type) []data.Type {
	sort.Slice(types, func(i, j int) bool {
		if types[i].Package == types[j].Package {
			return types[i].LocalType < types[j].LocalType
		}
		return types[i].Package < types[j].Package
	})
	return types
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
func addComponentToRow(row *bytes.Buffer, comp data.Type, partMap map[string]*sourcePart) {
	cNam := typeToString(comp)
	flow := partMap[markerFlow+cNam]
	fun := partMap[markerFunc+cNam]
	if flow != nil {
		// [link to Google!](http://google.com)
		row.WriteString(
			"[" + cNam + "](" +
				"./" + flow.mdFile.name + ".md#flow-" +
				strings.ToLower(flow.name) +
				")")
	} else if fun != nil {
		row.WriteString(fmt.Sprintf(
			"[%s](./%s#L%dL%d)",
			cNam, fun.goFile, fun.start, fun.end,
		))
	} else {
		row.WriteString(cNam)
	}
}
func addTypeToRow(row *bytes.Buffer, typ data.Type, partMap map[string]*sourcePart) {
	tNam := typeToString(typ)
	ty := partMap[markerType+tNam]

	if ty != nil {
		row.WriteString(fmt.Sprintf(
			"[%s](./%s#L%dL%d)",
			tNam, ty.goFile, ty.start, ty.end,
		))
	} else {
		row.WriteString(tNam)
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

func endMDFile(f *mdFile) error {
	if f == nil || f.osfile == nil {
		return nil
	}
	return f.osfile.Close()
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
