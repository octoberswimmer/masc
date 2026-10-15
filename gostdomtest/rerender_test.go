package gostdomtest

import (
	"strconv"
	"strings"
	"testing"

	"github.com/gost-dom/browser/dom"
	html "github.com/gost-dom/browser/html"

	"github.com/octoberswimmer/masc"
	"github.com/octoberswimmer/masc/elem"
)

type step struct{}

// host changes a text node on every message and shows a child between two
// siblings while its count is odd, which exercises the text, insertion, and
// removal paths of the gost-dom adapter on a re-render.
type host struct {
	masc.Core
	n int
}

func (h *host) Init() masc.Cmd { return nil }

func (h *host) Update(msg masc.Msg) (masc.Model, masc.Cmd) {
	if _, ok := msg.(step); ok {
		h.n++
	}
	return h, nil
}

func (h *host) Render(send func(masc.Msg)) masc.ComponentOrHTML {
	children := []masc.MarkupOrChild{
		masc.Markup(masc.Attribute("id", "host")),
		elem.Span(masc.Text("first")),
	}
	if h.n%2 == 1 {
		children = append(children, elem.Button(masc.Markup(masc.Attribute("id", "middle")), masc.Text("inserted")))
	}
	children = append(children, elem.Span(masc.Markup(masc.Attribute("id", "count")), masc.Text("count="+strconv.Itoa(h.n))))
	return elem.Div(children...)
}

func render(t *testing.T, steps int) html.Window {
	t.Helper()
	win, err := html.NewWindowReader(strings.NewReader("<!DOCTYPE html><html><body><div id=\"root\"></div></body></html>"))
	if err != nil {
		t.Fatalf("failed to create window: %v", err)
	}
	masc.UseGostDOM(win)
	root, err := win.Document().QuerySelector("#root")
	if err != nil || root == nil {
		t.Fatalf("root not found: %v", err)
	}
	program := masc.NewProgram(&host{}, masc.RenderTo(masc.WrapGostNode(root)))
	done := make(chan error)
	go func() {
		_, err := program.Run()
		done <- err
	}()
	for i := 0; i < steps; i++ {
		program.Send(step{})
	}
	program.Send(masc.Quit())
	if err := <-done; err != nil {
		t.Fatalf("program error: %v", err)
	}
	return win
}

func textOf(t *testing.T, win html.Window, selector string) string {
	t.Helper()
	el, err := win.Document().QuerySelector(selector)
	if err != nil || el == nil {
		t.Fatalf("%s not found: %v", selector, err)
	}
	return el.TextContent()
}

func TestRerenderUpdatesText(t *testing.T) {
	win := render(t, 2)
	if got := textOf(t, win, "#count"); got != "count=2" {
		t.Errorf("expected the text node to change on re-render, got %q", got)
	}
}

func TestRerenderInsertsBetweenSiblings(t *testing.T) {
	win := render(t, 1)
	hostEl, err := win.Document().QuerySelector("#host")
	if err != nil || hostEl == nil {
		t.Fatalf("host not found: %v", err)
	}
	var names []string
	for _, n := range hostEl.ChildNodes().All() {
		if el, ok := n.(dom.Element); ok {
			names = append(names, el.TagName())
		}
	}
	if strings.Join(names, ",") != "SPAN,BUTTON,SPAN" {
		t.Errorf("expected the button inserted between the spans, got %v", names)
	}
	if got := textOf(t, win, "#middle"); got != "inserted" {
		t.Errorf("inserted child text = %q", got)
	}
}

func TestRerenderRemovesAMiddleChild(t *testing.T) {
	win := render(t, 2)
	// n goes 0 -> 1 -> 2: the button is inserted and then removed.
	el, err := win.Document().QuerySelector("#middle")
	if err != nil {
		t.Fatal(err)
	}
	if el != nil {
		t.Error("expected the button removed on the second re-render")
	}
	if got := textOf(t, win, "#count"); got != "count=2" {
		t.Errorf("count = %q", got)
	}
}
