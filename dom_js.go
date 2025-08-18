//go:build js
// +build js

package masc

import (
	"fmt"
	"runtime/debug"
	"syscall/js"
)

// Event represents a DOM event.
type Event struct {
	js.Value
	Target js.Value
}

// Node returns the underlying JavaScript Element or TextNode.
//
// It panics if it is called before the DOM node has been attached, i.e. before
// the associated component's Mounter interface would be invoked.
func (h *HTML) Node() js.Value {
	if h.node == nil {
		panic("masc: cannot call (*HTML).Node() before DOM node creation / component mount")
	}
	return h.node.(wrappedObject).j
}

// RenderIntoNode renders the given component into the existing HTML element by
// replacing it.
//
// If the Component's Render method does not return an element of the same type,
// an error of type ElementMismatchError is returned.
func RenderIntoNode(node js.Value, c Component, send func(Msg)) error {
	return renderIntoNode("RenderIntoNode", wrapObject(node), c, send)
}

// RenderTo configures the renderer to render the model to the passed DOM node.
func RenderTo(rootNode js.Value) ProgramOption {
	return func(p *Program) {
		p.renderer = newNodeRenderer(wrapObject(rootNode))
	}
}

func toLower(s string) string {
	// We must call the prototype method here to workaround a limitation of
	// syscall/js in both Go and GopherJS where we cannot call the
	// `toLowerCase` string method. See https://github.com/golang/go/issues/35917
	return js.Global().Get("String").Get("prototype").Get("toLowerCase").Call("call", js.ValueOf(s)).String()
}

var globalValue jsObject

func global() jsObject {
	if globalValue == nil {
		globalValue = wrapObject(js.Global())
	}
	return globalValue
}

func undefined() wrappedObject {
	return wrappedObject{js.Undefined()}
}

func funcOf(fn func(this jsObject, args []jsObject) interface{}) jsFunc {
	return &jsFuncImpl{
		f: js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			wrappedArgs := make([]jsObject, len(args))
			for i, arg := range args {
				wrappedArgs[i] = wrapObject(arg)
			}
			return unwrap(fn(wrapObject(this), wrappedArgs))
		}),
		goFunc: fn,
	}
}

type jsFuncImpl struct {
	f      js.Func
	goFunc func(this jsObject, args []jsObject) interface{}
}

func (j *jsFuncImpl) String() string {
	// fmt.Sprint(j) would produce the actual implementation of the function in
	// JS code which differs across WASM/GopherJS/TinyGo so we instead just
	// return an opaque string for testing purposes.
	return "func"
}

func (j *jsFuncImpl) Release() { j.f.Release() }

func valueOf(v interface{}) jsObject {
	return wrapObject(js.ValueOf(v))
}

func wrapObject(j js.Value) jsObject {
	if j.IsNull() {
		return nil
	}
	return wrappedObject{j: j}
}

func unwrap(value interface{}) interface{} {
	if v, ok := value.(wrappedObject); ok {
		return v.j
	}
	if v, ok := value.(*jsFuncImpl); ok {
		return v.f
	}
	return value
}

type wrappedObject struct {
	j js.Value
}

func (w wrappedObject) Set(key string, value interface{}) {
	w.j.Set(key, unwrap(value))
}

func (w wrappedObject) Get(key string) jsObject {
	return wrapObject(w.j.Get(key))
}

func (w wrappedObject) Delete(key string) {
	w.j.Delete(key)
}

func (w wrappedObject) Call(name string, args ...interface{}) jsObject {
	for i, arg := range args {
		args[i] = unwrap(arg)
	}
	return wrapObject(w.j.Call(name, args...))
}

func (w wrappedObject) String() string {
	return w.j.String()
}

func (w wrappedObject) Truthy() bool {
	return w.j.Truthy()
}

func (w wrappedObject) IsUndefined() bool {
	return w.j.IsUndefined()
}

func (w wrappedObject) Equal(other jsObject) bool {
	if w.j.IsNull() != (other == nil) {
		return false
	}
	return w.j.Equal(unwrap(other).(js.Value))
}

func (w wrappedObject) Bool() bool {
	return w.j.Bool()
}

func (w wrappedObject) Int() int {
	return w.j.Int()
}

func (w wrappedObject) Float() float64 {
	return w.j.Float()
}

// requestAnimationFrame calls the native JS function of the same name.
func requestAnimationFrame(callback func(float64, func(Msg)), send func(Msg)) {
	var cb jsFunc
	cb = funcOf(func(_ jsObject, args []jsObject) interface{} {
		cb.Release()

		// Add panic recovery for render callbacks
		defer func() {
			if r := recover(); r != nil {
				js.Global().Get("console").Call("log", "MASC caught panic in render callback:", r)

				if currentProgram != nil {
					js.Global().Get("console").Call("log", "Calling panic handler")
					currentProgram.panicHandler(r)
				} else {
					// Fallback if no current program (shouldn't happen)
					js.Global().Get("console").Call("log", "No current program - using fallback")
					fmt.Printf("Caught panic in render callback:\n\n%s\n\n", r)
					debug.PrintStack()
				}
			}
		}()

		callback(args[0].Float(), send)
		return undefined()
	})
	global().Call("requestAnimationFrame", cb)
}

// reconcileProperties updates properties/attributes/etc to match the current
// element.
func (h *HTML) reconcileProperties(prev *HTML) {
	// If nodes match, remove any outdated properties
	if h.node.Equal(prev.node) {
		h.removeProperties(prev)
	}

	// Wrap event listeners
	for _, l := range h.eventListeners {
		l := l
		l.wrapper = funcOf(func(_ jsObject, args []jsObject) interface{} {
			jsEvent := args[0]
			if l.callPreventDefault {
				jsEvent.Call("preventDefault")
			}
			if l.callStopPropagation {
				jsEvent.Call("stopPropagation")
			}
			l.Listener(&Event{
				Value:  jsEvent.(wrappedObject).j,
				Target: jsEvent.Get("target").(wrappedObject).j,
			})
			return undefined()
		})
	}

	// Properties
	for name, value := range h.properties {
		var oldValue interface{}
		switch name {
		case "value":
			oldValue = h.node.Get("value").String()
		case "checked":
			oldValue = h.node.Get("checked").Bool()
		default:
			oldValue = prev.properties[name]
		}
		if value != oldValue {
			h.node.Set(name, value)
		}
	}

	// Attributes
	for name, value := range h.attributes {
		if value != prev.attributes[name] {
			h.node.Call("setAttribute", name, value)
		}
	}

	// Classes
	classList := h.node.Get("classList")
	for name := range h.classes {
		if _, ok := prev.classes[name]; !ok {
			classList.Call("add", name)
		}
	}

	// Dataset
	dataset := h.node.Get("dataset")
	for name, value := range h.dataset {
		if value != prev.dataset[name] {
			dataset.Set(name, value)
		}
	}

	// Styles
	style := h.node.Get("style")
	for name, value := range h.styles {
		oldValue := prev.styles[name]
		if value != oldValue {
			style.Call("setProperty", name, value)
		}
	}

	// Event listeners
	for _, l := range h.eventListeners {
		h.node.Call("addEventListener", l.Name, l.wrapper)
	}

	// InnerHTML
	if h.innerHTML != prev.innerHTML {
		h.node.Set("innerHTML", h.innerHTML)
	}
}
