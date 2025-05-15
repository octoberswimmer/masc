package example

import (
	"strings"
	"testing"

	html "github.com/gost-dom/browser/html"
	"github.com/octoberswimmer/masc"
	"github.com/octoberswimmer/masc/elem"
)

// Define a component that tracks mount/unmount calls.
type lifeComp struct {
	masc.Core
	onMount   *int
	onUnmount *int
}

// Implement lifecycle methods.
func (l *lifeComp) Init() masc.Cmd                             { return nil }
func (l *lifeComp) Update(msg masc.Msg) (masc.Model, masc.Cmd) { return l, nil }
func (l *lifeComp) Render(send func(masc.Msg)) masc.ComponentOrHTML {
	return elem.Body(elem.Div())
}
func (l *lifeComp) Mount() {
	if l.onMount != nil {
		*l.onMount++
	}
}
func (l *lifeComp) Unmount() {
	if l.onUnmount != nil {
		*l.onUnmount++
	}
}

// Parent toggles child off after first render.
type parentComp struct {
	masc.Core
	child *lifeComp
}

func (p *parentComp) Init() masc.Cmd { return nil }
func (p *parentComp) Update(msg masc.Msg) (masc.Model, masc.Cmd) {
	p.child = nil
	return p, nil
}
func (p *parentComp) Render(send func(masc.Msg)) masc.ComponentOrHTML {
	if p.child != nil {
		return p.child
	}
	return elem.Body(elem.Div())
}

// Component that always skips its re-render, rendering a <div> with a data-count attribute.
type skipComp struct {
	masc.Core
	count int
}

func (c *skipComp) Init() masc.Cmd                             { return nil }
func (c *skipComp) Update(msg masc.Msg) (masc.Model, masc.Cmd) { c.count++; return c, nil }
func (c *skipComp) SkipRender(prev masc.Component) bool        { return true }

// Render returns a <div> element with the current count; always using SkipRender to prevent re-render.
func (c *skipComp) Render(send func(masc.Msg)) masc.ComponentOrHTML {
	return elem.Div(masc.Markup(masc.Property("data-count", c.count)))
}

// TestLifecycleHooks verifies Mount and Unmount on components.
func TestLifecycleHooks(t *testing.T) {
	win, err := html.NewWindowReader(strings.NewReader("<!DOCTYPE html><html><body></body></html>"))
	if err != nil {
		t.Fatalf("failed to create gost-dom window: %v", err)
	}
	masc.UseGostDOM(win)
	mountCount, unmountCount := 0, 0

	child := &lifeComp{onMount: &mountCount, onUnmount: &unmountCount}
	parent := &parentComp{child: child}
	_, send, err := masc.RenderComponentIntoWithSend(win, parent)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if mountCount != 1 {
		t.Errorf("expected mountCount=1, got %d", mountCount)
	}
	if unmountCount != 0 {
		t.Errorf("expected unmountCount=0, got %d", unmountCount)
	}
	// Trigger update to remove child.
	send(nil)
	if unmountCount != 1 {
		t.Errorf("expected unmountCount=1 after removal, got %d", unmountCount)
	}
}

func TestDatasetAndStyleProxy(t *testing.T) {
	win, err := html.NewWindowReader(strings.NewReader("<!DOCTYPE html><html><body></body></html>"))
	if err != nil {
		t.Fatalf("failed to create gost-dom window: %v", err)
	}
	// Configure masc to use gost-dom on this window
	masc.UseGostDOM(win)
	// Wrap the <body> element for jsObject operations
	root := masc.WrapGostNode(win.Document().Body())

	// --- Dataset proxy ---
	ds := root.Get("dataset")
	ds.Set("foo", "bar")
	val, ok := win.Document().Body().GetAttribute("data-foo")
	if !ok || val != "bar" {
		t.Errorf("dataset.Set failed: got %q, want %q", val, "bar")
	}
	ds.Delete("foo")
	if _, ok := win.Document().Body().GetAttribute("data-foo"); ok {
		t.Errorf("dataset.Delete failed: data-foo still present")
	}

	// --- Style proxy ---
	style := root.Get("style")
	style.Call("setProperty", "background", "blue")
	css, _ := win.Document().Body().GetAttribute("style")
	if !strings.Contains(css, "background:blue") {
		t.Errorf("style.setProperty failed: got %q, want to contain %q", css, "background:blue")
	}
	style.Call("removeProperty", "background")
	css2, _ := win.Document().Body().GetAttribute("style")
	if strings.Contains(css2, "background:blue") {
		t.Errorf("style.removeProperty failed: got %q, still contains %q", css2, "background:blue")
	}
}

// TestSkipRender ensures SkipRender prevents DOM updates when nothing changes
// for components rendered into a <div>. SkipRender is not supported when rendering to <body>.
func TestSkipRender(t *testing.T) {
	win, err := html.NewWindowReader(strings.NewReader("<!DOCTYPE html><html><body></body></html>"))
	if err != nil {
		t.Fatalf("failed to create gost-dom window: %v", err)
	}
	masc.UseGostDOM(win)

	// Create a <div> to render the skipComp into
	doc := win.Document()
	div := doc.CreateElement("div")
	// Append the div to the document body
	if _, err := doc.Body().AppendChild(div); err != nil {
		t.Fatalf("failed to append div to body: %v", err)
	}
	// Wrap the div node and prepare the component as a Masc program targeting the div
	comp := &skipComp{}
	program := masc.NewProgram(comp, masc.RenderTo(masc.WrapGostNode(div)))
	// Run the program in a separate goroutine
	done := make(chan struct{})
	var runErr error
	go func() {
		_, runErr = program.Run()
		close(done)
	}()
	// Send update messages; SkipRender should prevent attribute changes
	program.Send(nil)
	program.Send(nil)
	program.Send(nil)
	program.Send(nil)
	// Signal the program to quit and wait for it to finish
	program.Send(masc.Quit())
	<-done
	if runErr != nil {
		t.Fatalf("program run error: %v", runErr)
	}
	// Query the rendered <div> from the document and inspect its data-count
	el, err := win.Document().QuerySelector("div")
	if err != nil {
		t.Fatalf("query selector error: %v", err)
	}
	if el == nil {
		t.Fatal("expected a <div> element in document, found none")
	}
	val, ok := el.GetAttribute("data-count")
	if !ok {
		t.Fatalf("expected data-count attribute on <div>, got none")
	}
	if val != "0" {
		t.Errorf("SkipRender failed: expected data-count=0, got %q", val)
	}
}
