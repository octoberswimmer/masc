//go:build js

package masc

import "syscall/js"

// NewObject constructs a JavaScript object with the provided properties.
func NewObject(props map[string]interface{}) SyscallJSValue {
	obj := js.Global().Get("Object").New()
	for k, v := range props {
		obj.Set(k, v)
	}
	return obj
}
