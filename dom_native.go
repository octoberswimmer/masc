//go:build !js
// +build !js

package masc

import (
	"fmt"
	"strings"

	dom "github.com/gost-dom/browser/dom"
	ev "github.com/gost-dom/browser/dom/event"
	html "github.com/gost-dom/browser/html"
)

// Stubs for building masc under a native GOOS and GOARCH, so that masc
// type-checks, lints, auto-completes, and serves documentation under godoc.org
// as with any other normal Go package that is not under GOOS=js and
// GOARCH=wasm.

// SyscallJSValue is an actual syscall/js.Value type under WebAssembly compilation.
//
// It is declared here just for purposes of testing masc under native
// 'go test', linting, and serving documentation under godoc.org.
type SyscallJSValue jsObject

// Event represents a DOM event.
type Event struct {
	Value  SyscallJSValue
	Target SyscallJSValue
}

// gostPerformance implements jsObject for performance.now()
type gostPerformance struct{}

func (p *gostPerformance) Get(string) jsObject     { return nil }
func (p *gostPerformance) Set(string, interface{}) {}
func (p *gostPerformance) Delete(string)           {}
func (p *gostPerformance) Call(name string, _ ...interface{}) jsObject {
	if name == "now" {
		return &floatObject{f: 0}
	}
	panic("gostdom: performance.Call(\"" + name + "\") not implemented")
}
func (p *gostPerformance) String() string        { return "[performance]" }
func (p *gostPerformance) Truthy() bool          { return true }
func (p *gostPerformance) Equal(o jsObject) bool { _, ok := o.(*gostPerformance); return ok }
func (p *gostPerformance) IsUndefined() bool     { return false }
func (p *gostPerformance) Bool() bool            { return true }
func (p *gostPerformance) Int() int              { return 0 }
func (p *gostPerformance) Float() float64        { return 0 }

// floatObject implements jsObject for numeric calls
type floatObject struct{ f float64 }

func (f *floatObject) Get(string) jsObject                  { return nil }
func (f *floatObject) Set(string, interface{})              {}
func (f *floatObject) Delete(string)                        {}
func (f *floatObject) Call(string, ...interface{}) jsObject { return nil }
func (f *floatObject) String() string                       { return fmt.Sprint(f.f) }
func (f *floatObject) Truthy() bool                         { return f.f != 0 }
func (f *floatObject) Equal(o jsObject) bool {
	if o2, ok := o.(*floatObject); ok {
		return f.f == o2.f
	}
	return false
}
func (f *floatObject) IsUndefined() bool { return false }
func (f *floatObject) Bool() bool        { return f.f != 0 }
func (f *floatObject) Int() int          { return int(f.f) }
func (f *floatObject) Float() float64    { return f.f }

// requestAnimationFrame schedules the callback via global.Call, as in JS.
// This allows tests to intercept and record the requestAnimationFrame call.
func requestAnimationFrame(callback func(float64, func(Msg)), send func(Msg)) {
	var cb jsFunc
	cb = funcOf(func(_ jsObject, args []jsObject) interface{} {
		cb.Release()
		callback(args[0].Float(), send)
		return undefined()
	})
	global().Call("requestAnimationFrame", cb)
}

// Node returns the underlying JavaScript Element or TextNode.
//
// It panics if it is called before the DOM node has been attached, i.e. before
// the associated component's Mounter interface would be invoked.
func (h *HTML) Node() SyscallJSValue {
	return htmlNodeImpl(h)
}

// RenderIntoNode renders the given component into the existing HTML element by
// replacing it.
//
// If the Component's Render method does not return an element of the same type,
// an error of type ElementMismatchError is returned.
func RenderIntoNode(node SyscallJSValue, c Component, send func(Msg)) error {
	return renderIntoNode("RenderIntoNode", node, c, send)
}

// RenderTo configures the renderer to render the model to the passed DOM node.
func RenderTo(rootNode SyscallJSValue) ProgramOption {
	return func(p *Program) {
		p.renderer = newNodeRenderer(rootNode)
	}
}

func toLower(s string) string {
	return strings.ToLower(s)
}

var globalValue jsObject

func global() jsObject {
	return globalValue
}

func undefined() wrappedObject {
	return wrappedObject{j: &jsObjectImpl{}}
}

func funcOf(fn func(this jsObject, args []jsObject) interface{}) jsFunc {
	return funcOfImpl(fn)
}

type wrappedObject struct {
	jsObject
	j jsObject
}

type jsObjectImpl struct {
	jsObject
}

func (e *jsObjectImpl) Equal(other jsObject) bool {
	return e == other.(*jsObjectImpl)
}

var (
	htmlNodeImpl = func(_ *HTML) SyscallJSValue {
		panic("not implemented on this architecture in non-testing environment")
	}
	funcOfImpl = func(_ func(this jsObject, args []jsObject) interface{}) jsFunc {
		panic("not implemented on this architecture in non-testing environment")
	}
)

// UseGostDOM configures masc to render components into a gost-dom browser window.
// Call this before rendering when running under native Go (non-wasm).
func UseGostDOM(win html.Window) {
	globalValue = &gostGlobal{win: win}
	htmlNodeImpl = func(h *HTML) SyscallJSValue {
		if h.node == nil {
			panic("masc: (*HTML).Node() before DOM node creation")
		}
		return SyscallJSValue(h.node)
	}
	funcOfImpl = func(fn func(this jsObject, args []jsObject) interface{}) jsFunc {
		return &gostFunc{goFunc: fn}
	}
}

// WrapGostNode converts a gost-dom Node into a SyscallJSValue for rendering.
func WrapGostNode(n dom.Node) SyscallJSValue {
	return SyscallJSValue(&gostWrapper{n: n})
}

// gostGlobal implements jsObject over a gost-dom browser window.
type gostGlobal struct {
	win html.Window
}

func (g *gostGlobal) Get(key string) jsObject {
	switch key {
	case "document":
		return &gostWrapper{n: g.win.Document()}
	case "readyState":
		return &stringObject{s: "complete"}
	case "performance":
		return &gostPerformance{}
	}
	panic("gostdom: global.Get(\"" + key + "\") not implemented")
}
func (*gostGlobal) Set(_ string, _ interface{})              {}
func (*gostGlobal) Delete(_ string)                          {}
func (*gostGlobal) Call(_ string, _ ...interface{}) jsObject { return nil }
func (*gostGlobal) String() string                           { return "[gostdom global]" }
func (*gostGlobal) Truthy() bool                             { return true }
func (g *gostGlobal) Equal(o jsObject) bool                  { return g == o.(*gostGlobal) }
func (*gostGlobal) IsUndefined() bool                        { return false }
func (*gostGlobal) Bool() bool                               { return false }
func (*gostGlobal) Int() int                                 { return 0 }
func (*gostGlobal) Float() float64                           { return 0 }

// gostWrapper wraps a gost-dom/browser dom.Node and implements jsObject.
type gostWrapper struct {
	n dom.Node
}

func (g *gostWrapper) Set(key string, value interface{}) {
	switch key {
	case "innerHTML":
		if el, ok := g.n.(dom.Element); ok {
			// ignore error
			_ = el.SetInnerHTML(value.(string))
		}
	case "nodeValue":
		g.n.SetTextContent(value.(string))
	default:
		if el, ok := g.n.(dom.Element); ok {
			el.SetAttribute(key, fmt.Sprint(value))
		}
	}
}
func (g *gostWrapper) Get(key string) jsObject {
	switch key {
	case "innerHTML":
		if el, ok := g.n.(dom.Element); ok {
			txt := el.InnerHTML()
			node := g.n.OwnerDocument().CreateText(txt)
			return &gostWrapper{n: node}
		}
	case "nodeName":
		return &stringObject{s: g.n.NodeName()}
	case "parentNode":
		p := g.n.Parent()
		if p == nil {
			return nil
		}
		return &gostWrapper{n: p}
	case "readyState":
		return &stringObject{s: "complete"}
	case "value":
		if el, ok := g.n.(dom.Element); ok {
			val, _ := el.GetAttribute("value")
			return &stringObject{s: val}
		}
	case "classList":
		// Return the element itself as a pseudo-classList for native support
		return &gostWrapper{n: g.n}
	}
	if el, ok := g.n.(dom.Element); ok {
		val, _ := el.GetAttribute(key)
		node := g.n.OwnerDocument().CreateText(val)
		return &gostWrapper{n: node}
	}
	return nil
}
func (g *gostWrapper) Delete(key string) {
	switch key {
	case "innerHTML":
		if el, ok := g.n.(dom.Element); ok {
			// ignore error
			_ = el.SetInnerHTML("")
		}
	default:
		if el, ok := g.n.(dom.Element); ok {
			el.RemoveAttribute(key)
		}
	}
}
func (g *gostWrapper) Call(name string, args ...interface{}) jsObject {
	switch name {
	case "add":
		// Handle classList.add on native: add CSS class by updating class attribute
		if el, ok := g.n.(dom.Element); ok {
			cls := args[0].(string)
			current, _ := el.GetAttribute("class")
			updated := current
			if updated != "" {
				updated += " "
			}
			updated += cls
			el.SetAttribute("class", updated)
		}
		return nil
	case "appendChild":
		child := args[0].(*gostWrapper).n
		// ignore returned node and error
		_, _ = g.n.AppendChild(child)
		return args[0].(*gostWrapper)
	case "removeChild":
		child := args[0].(*gostWrapper).n
		// ignore returned node and error
		_, _ = g.n.RemoveChild(child)
		return args[0].(*gostWrapper)
	case "replaceChild":
		child := args[0].(*gostWrapper).n
		old := args[1].(*gostWrapper).n
		// ignore returned node and error
		_, _ = g.n.InsertBefore(child, old)
		_, _ = g.n.RemoveChild(old)
		return &gostWrapper{n: child}
	case "insertBefore":
		child := args[0].(*gostWrapper).n
		ref := args[1].(*gostWrapper).n
		// ignore returned node and error
		_, _ = g.n.InsertBefore(child, ref)
		return args[0].(*gostWrapper)
	case "createElement":
		tag := args[0].(string)
		if docNode, ok := g.n.(dom.Document); ok {
			el := docNode.CreateElement(tag)
			return &gostWrapper{n: el}
		}
		el := g.n.OwnerDocument().CreateElement(tag)
		return &gostWrapper{n: el}
	case "createElementNS":
		ns := args[0].(string)
		tag := args[1].(string)
		if docNode, ok := g.n.(dom.Document); ok {
			el := docNode.CreateElementNS(ns, tag)
			return &gostWrapper{n: el}
		}
		el := g.n.OwnerDocument().CreateElementNS(ns, tag)
		return &gostWrapper{n: el}
	case "createTextNode":
		txt := args[0].(string)
		var node dom.Text
		if docNode, ok := g.n.(dom.Document); ok {
			node = docNode.CreateText(txt)
		} else {
			node = g.n.OwnerDocument().CreateText(txt)
		}
		return &gostWrapper{n: node}
	case "setAttribute":
		if el, ok := g.n.(dom.Element); ok {
			key := args[0].(string)
			val := fmt.Sprint(args[1])
			el.SetAttribute(key, val)
		}
		return nil
	case "removeAttribute":
		if el, ok := g.n.(dom.Element); ok {
			key := args[0].(string)
			el.RemoveAttribute(key)
		}
		return nil
	case "addEventListener":
		if el, ok := g.n.(ev.EventTarget); ok {
			eventType := args[0].(string)
			if cb, ok2 := args[1].(*gostFunc); ok2 {
				handler := ev.NewEventHandlerFuncWithoutError(func(evt *ev.Event) {
					// Wrap the gost-dom Event for user callback
					ge := &gostEvent{ev: evt}
					// Invoke callback with the event wrapper
					cb.goFunc(ge, []jsObject{ge})
				})
				el.AddEventListener(eventType, handler)
			}
		}
		return nil
	case "removeEventListener":
		if el, ok := g.n.(ev.EventTarget); ok {
			eventType := args[0].(string)
			if cb, ok2 := args[1].(*gostFunc); ok2 {
				handler := ev.NewEventHandlerFuncWithoutError(func(evt *ev.Event) {
					if tgtNode, ok3 := evt.Target().(dom.Node); ok3 {
						w := &wrappedObject{j: &gostWrapper{n: tgtNode}}
						cb.goFunc(w, []jsObject{w})
					}
				})
				el.RemoveEventListener(eventType, handler)
			}
		}
		return nil
	case "querySelector":
		// CSS selector single element
		if len(args) > 0 {
			if sel, ok := args[0].(string); ok {
				switch n := g.n.(type) {
				case dom.Document:
					if el, _ := n.QuerySelector(sel); el != nil {
						return &gostWrapper{n: el}
					}
				case dom.Element:
					if el, _ := n.QuerySelector(sel); el != nil {
						return &gostWrapper{n: el}
					}
				}
			}
		}
		return nil
	case "querySelectorAll":
		// CSS selector matching elements
		if len(args) > 0 {
			if sel, ok := args[0].(string); ok {
				var list dom.NodeList
				switch n := g.n.(type) {
				case dom.Document:
					if l, _ := n.QuerySelectorAll(sel); l != nil {
						list = l
					}
				case dom.Element:
					if l, _ := n.QuerySelectorAll(sel); l != nil {
						list = l
					}
				}
				if list != nil {
					return &gostNodeList{list: list}
				}
			}
		}
		return &gostNodeList{list: nil}
	case "dispatchEvent":
		// Dispatch a gost-dom Event
		if tgt, ok := g.n.(ev.EventTarget); ok {
			if ge, ok2 := args[0].(*gostEvent); ok2 && ge.ev != nil {
				tgt.DispatchEvent(ge.ev)
			}
		}
		return nil
	}
	panic("gostdom: Call \"" + name + "\" not implemented")
}
func (g *gostWrapper) String() string { return g.n.NodeName() }
func (g *gostWrapper) Truthy() bool   { return true }
func (g *gostWrapper) Equal(o jsObject) bool {
	other, ok := o.(*gostWrapper)
	return ok && g.n == other.n
}
func (g *gostWrapper) IsUndefined() bool { return false }
func (*gostWrapper) Bool() bool          { return false }
func (*gostWrapper) Int() int            { return 0 }
func (*gostWrapper) Float() float64      { return 0 }

// gostFunc implements jsFunc for event callbacks (no-op Release).
type gostFunc struct {
	goFunc func(this jsObject, args []jsObject) interface{}
}

func (*gostFunc) Release() {}

// stringObject wraps a Go string as a jsObject.
type stringObject struct{ s string }

func (s *stringObject) String() string { return s.s }
func (s *stringObject) Truthy() bool   { return true }
func (s *stringObject) Equal(o jsObject) bool {
	if o2, ok := o.(*stringObject); ok {
		return s.s == o2.s
	}
	return false
}
func (*stringObject) Set(string, interface{})              {}
func (*stringObject) Get(string) jsObject                  { return nil }
func (*stringObject) Delete(string)                        {}
func (*stringObject) Call(string, ...interface{}) jsObject { return nil }
func (*stringObject) IsUndefined() bool                    { return false }
func (*stringObject) Bool() bool                           { return false }
func (*stringObject) Int() int                             { return 0 }
func (*stringObject) Float() float64                       { return 0 }

// gostNodeList wraps a gost-dom NodeList and implements jsObject.
// It provides length and item access.
type gostNodeList struct {
	list dom.NodeList
}

func (g *gostNodeList) Set(key string, value interface{}) {}
func (g *gostNodeList) Get(key string) jsObject {
	switch key {
	case "length":
		return &floatObject{f: float64(g.list.Length())}
	}
	return nil
}
func (g *gostNodeList) Delete(key string) {}
func (g *gostNodeList) Call(name string, args ...interface{}) jsObject {
	switch name {
	case "item":
		if len(args) > 0 {
			if idx, ok := args[0].(int); ok {
				node := g.list.Item(idx)
				if node != nil {
					return &gostWrapper{n: node}
				}
			}
		}
	}
	return nil
}
func (g *gostNodeList) String() string { return "[gost-dom nodelist]" }
func (g *gostNodeList) Truthy() bool   { return true }
func (g *gostNodeList) Equal(o jsObject) bool {
	if o2, ok := o.(*gostNodeList); ok {
		return g == o2
	}
	return false
}
func (g *gostNodeList) IsUndefined() bool { return false }
func (g *gostNodeList) Bool() bool        { return false }
func (g *gostNodeList) Int() int          { return g.list.Length() }
func (g *gostNodeList) Float() float64    { return float64(g.list.Length()) }

// gostEvent wraps a gost-dom Event so it can be passed through dispatchEvent.
// It implements jsObject over ev.Event.
type gostEvent struct {
	ev *ev.Event
}

func (e *gostEvent) Set(key string, value interface{}) {}
func (e *gostEvent) Delete(key string)                 {}

func (e *gostEvent) Get(key string) jsObject {
	if key == "value" {
		if el, ok := e.ev.Target().(dom.Element); ok {
			val, _ := el.GetAttribute("value")
			return &stringObject{s: val}
		}
	}
	return nil
}

func (e *gostEvent) Call(name string, _ ...interface{}) jsObject {
	switch name {
	case "preventDefault":
		e.ev.PreventDefault()
	case "stopPropagation":
		e.ev.StopPropagation()
	}
	return undefined()
}

func (e *gostEvent) String() string        { return "[gostdom event]" }
func (e *gostEvent) Truthy() bool          { return true }
func (e *gostEvent) Equal(o jsObject) bool { _, ok := o.(*gostEvent); return ok }
func (e *gostEvent) IsUndefined() bool     { return false }
func (e *gostEvent) Bool() bool            { return true }
func (e *gostEvent) Int() int              { return 0 }
func (e *gostEvent) Float() float64        { return 0 }

// WithGostDOM configures masc to render into a gost-dom Window in native builds.
// It calls UseGostDOM under the hood.
func WithGostDOM(win html.Window) ProgramOption {
	return func(p *Program) {
		UseGostDOM(win)
	}
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
			// Value is the event wrapper, Target is jsEvent.Get("target")
			valObj := jsEvent
			tgtObj := jsEvent.Get("target")
			l.Listener(&Event{
				Value:  SyscallJSValue(valObj),
				Target: SyscallJSValue(tgtObj),
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
