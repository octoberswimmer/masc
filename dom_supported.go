//go:build !js
// +build !js

package masc

import (
	"flag"
	"os"
	"strings"
)

func init() {
	// skip browser guard when running under `go test` or other test binaries
	if isTest || flag.Lookup("test.v") != nil || (len(os.Args) > 0 && (strings.HasSuffix(os.Args[0], ".test") || strings.HasSuffix(os.Args[0], ".test.exe"))) {
		return
	}
	if global() == nil {
		panic("masc: only WebAssembly and testing compilation is supported")
	}
	if global().Get("document").IsUndefined() {
		panic("masc: only running inside a browser is supported")
	}
}
