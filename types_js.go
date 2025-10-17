//go:build js

package masc

import "syscall/js"

// SyscallJSValue mirrors syscall/js.Value when building for WebAssembly.
type SyscallJSValue = js.Value
