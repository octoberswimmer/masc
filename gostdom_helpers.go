//go:build !js
// +build !js

package masc

import (
	"fmt"

	ev "github.com/gost-dom/browser/dom/event"
	"github.com/gost-dom/browser/html"
)

// RenderComponentInto renders the given Model into the <body> of the provided gost-dom Window.
// It configures masc to use gost-dom, performs the initial render, and wires subsequent re-renders
// automatically upon messages (e.g., from event listeners).
// Returns a Body handle for inspection.
// Body proxies interactions to the current document body.
type Body struct {
	win html.Window
}

// Click dispatches a click event on the document body.
func (b Body) Click() {
	if elt, ok := b.win.Document().Body().(html.HTMLElement); ok {
		elt.Click()
	}
}

// InnerHTML returns the current innerHTML of the document body.
func (b Body) InnerHTML() string {
	return b.win.Document().Body().InnerHTML()
}

// Dispatch dispatches a DOM event of the given type on the element matching selector.
// It uses WrapGostNode and the gostWrapper "dispatchEvent" case.
func (b Body) Dispatch(selector, eventType string) error {
	node, err := b.win.Document().QuerySelector(selector)
	if err != nil {
		return fmt.Errorf("query selector %q error: %w", selector, err)
	}
	// Create a gost-dom Event and dispatch
	ge := &gostEvent{ev: &ev.Event{Type: eventType}}
	WrapGostNode(node).Call("dispatchEvent", ge)
	return nil
}
func RenderComponentInto(win html.Window, m Model) (Body, error) {
	// Configure masc to use gost-dom via Window
	UseGostDOM(win)
	// Obtain the document and <body> element
	doc := win.Document()
	body := doc.Body()
	if body == nil {
		return Body{}, fmt.Errorf("gostdom: <body> element not found")
	}
	// Wrap for masc
	root := WrapGostNode(body)
	// Local model copy
	model := m
	// Send function triggers re-render on the current <body>
	var send func(Msg)
	send = func(msg Msg) {
		updated, _ := model.Update(msg)
		model = updated
		// Re-query the <body> element and wrap as root
		body = win.Document().Body()
		root = WrapGostNode(body)
		_ = RenderIntoNode(root, model, send)
	}
	// Initial render
	if err := RenderIntoNode(root, model, send); err != nil {
		return Body{}, err
	}
	// After render, the <body> element may have been replaced, re-query it
	// Return a body handle for tests
	return Body{win: win}, nil
}

// RenderComponentIntoWithSend renders the given Model into the <body> of the provided gost-dom Window.
// It returns a Body handle and the send function used for dispatching messages.
func RenderComponentIntoWithSend(win html.Window, m Model) (Body, func(Msg), error) {
	UseGostDOM(win)
	doc := win.Document()
	bodyNode := doc.Body()
	if bodyNode == nil {
		return Body{}, nil, fmt.Errorf("gostdom: <body> element not found")
	}
	root := WrapGostNode(bodyNode)
	model := m
	var send func(Msg)
	send = func(msg Msg) {
		updated, _ := model.Update(msg)
		model = updated
		// Re-query and wrap the current <body> element before re-render
		bodyNode := win.Document().Body()
		root = WrapGostNode(bodyNode)
		_ = RenderIntoNode(root, model, send)
	}
	if err := RenderIntoNode(root, model, send); err != nil {
		return Body{}, nil, err
	}
	return Body{win: win}, send, nil
}
