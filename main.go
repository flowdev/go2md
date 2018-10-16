package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flowdev/go2md/goast"
)

func main() {
	// find all the possible roots for Go source code in the right order:
	srcRoots := make([]string, 0, 4)
	vendorRoot := findVendorRoot()
	if vendorRoot != "" {
		srcRoots = append(srcRoots, vendorRoot)
	}
	srcRoots = append(srcRoots, findGoPathRoots()...)
	goRootRoot := findGoRootRoot()
	if goRootRoot != "" {
		srcRoots = append(srcRoots, goRootRoot)
	}
	fmt.Println("srcRoots:", srcRoots)
	if err := goast.ProcessDir("."); err != nil {
		log.Printf("FATAL: Unable to process current directory: %v", err)
	}
}

// findGoPathRoots finds all paths of GOPATH and turns them into source roots.
func findGoPathRoots() []string {
	gopath := getOutputOfCmd("go", "env", "GOPATH")
	gopaths := filepath.SplitList(gopath)
	srcRoots := make([]string, len(gopaths))
	for i, gp := range gopaths {
		srcRoots[i] = filepath.Join(gp, "src")
	}
	return srcRoots
}
func findVendorRoot() string {
	return crawlUpDirsAndFind("vendor", ".")
}
func findGoRootRoot() string {
	return filepath.Join(getOutputOfCmd("go", "env", "GOROOT"), "src")
}

func getOutputOfCmd(cmd string, args ...string) string {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		log.Fatalf("FATAL: Unable to execute command: %v", err)
	}
	return strings.TrimRight(string(out), "\r\n")
}
func crawlUpDirsAndFind(file, startDir string) string {
	absDir, err := filepath.Abs(startDir)
	if err != nil {
		log.Fatalf("FATAL: Unable to find absolute directory: %v", err)
	}
	volName := filepath.VolumeName(absDir)
	oldDir := "" // set to impossible value first!

	for ; absDir != volName && absDir != oldDir; absDir = filepath.Dir(absDir) {
		path := filepath.Join(absDir, file)
		if _, err = os.Stat(path); err == nil {
			return path
		}
		oldDir = absDir
	}
	return ""
}
