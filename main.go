package main

import (
	"log"

	"github.com/flowdev/go2md/goast"
)

func main() {
	if err := goast.ProcessDir("."); err != nil {
		log.Printf("FATAL: Unable to process current directory: %v", err)
	}
}

func findGoPath() string {
	// call 'go env GOPATH'
	// shall we support multiple paths?
	return ""
}
func findGoMod() string {
	// call 'go env GOMOD'
	return ""
}
func findVendor(goMod string) string {
	// crawl up directory structure or goMod
	return ""
}
func findGoRoot() string {
	// call 'go env GOROOT'
	return ""
}
