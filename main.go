package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flowdev/go2md/goast"
)

var localLinks bool

func init() {
	const (
		localLinksDefault = false
		localLinksUsage   = "create links to local files in markdown"
	)
	flag.BoolVar(&localLinks, "local", localLinksDefault, localLinksUsage)
	flag.BoolVar(&localLinks, "l", localLinksDefault, localLinksUsage+" (shorthand)")
}

func main() {
	flag.Parse()
	srcRoots := findSourceRoots()
	projRoot := getOutputOfCmd("git", "rev-parse", "--show-toplevel")
	fmt.Println("srcRoots:", srcRoots)
	fmt.Println("localLinks:", localLinks)
	fmt.Println("projRoot:", projRoot)
	if err := goast.ProcessDir(".", goast.NewPackageDict(srcRoots, projRoot, localLinks)); err != nil {
		log.Printf("FATAL: Unable to process current directory: %v", err)
	}
}

// findSourceRoots finds all the possible roots for Go source code in the right
// order.
func findSourceRoots() []string {
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
	return srcRoots
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
		log.Fatalf("FATAL: Unable to find absolute directory (for %s): %v", startDir, err)
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
