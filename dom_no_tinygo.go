//go:build !tinygo
// +build !tinygo

package masc

import (
	"flag"
	"os"
	"strings"
)

func init() {
	// skip browser/TinyGo guard when running under `go test` or other test binaries
	if isTest || flag.Lookup("test.v") != nil || (len(os.Args) > 0 && (strings.HasSuffix(os.Args[0], ".test") || strings.HasSuffix(os.Args[0], ".test.exe"))) {
		return
	}
	if global() == nil {
		panic("masc: only WebAssembly, TinyGo, and testing compilation is supported")
	}
	if global().Get("document").IsUndefined() {
		panic("masc: only running inside a browser is supported")
	}
}

func (h *HTML) tinyGoCannotIterateNilMaps() {}

func tinyGoAssertCopier(_ Component) {}
