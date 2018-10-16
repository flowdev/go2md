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
	fmt.Println("vendor:", findVendor())
	fmt.Println("gopath(s):", findGoPaths())
	fmt.Println("goroot:", findGoRoot())
	if err := goast.ProcessDir("."); err != nil {
		log.Printf("FATAL: Unable to process current directory: %v", err)
	}
}

// findGoPaths finds all paths of GOPATH and turns them into source roots.
func findGoPaths() []string {
	gopath := getOutputOfCmd("go", "env", "GOPATH")
	gopaths := filepath.SplitList(gopath)
	srcRoots := make([]string, len(gopaths))
	for i, gp := range gopaths {
		srcRoots[i] = filepath.Join(gp, "src")
	}
	return srcRoots
}
func findVendor() string {
	return crawlUpDirsAndFind("vendor", ".")
}
func findGoRoot() string {
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
		_, err := os.Lstat(path)
		if err == nil {
			return path
		}
		oldDir = absDir
	}
	return ""
}
