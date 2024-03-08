//go:build !tinygo
// +build !tinygo

package masc

func init() {
	if isTest {
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
