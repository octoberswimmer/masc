package main

import (
	"strings"
	"testing"

	"github.com/gost-dom/browser/html"
	"github.com/octoberswimmer/masc"
)

// TestPageViewInitialRender verifies that PageView renders correctly
// using the RenderComponentInto helper.
func TestPageViewInitialRender(t *testing.T) {
	win, err := html.NewWindowReader(strings.NewReader("<!DOCTYPE html><html><body></body></html>"))
	if err != nil {
		t.Fatalf("failed to create gost-dom window: %v", err)
	}
	bodyEl, err := masc.RenderComponentInto(win, &PageView{input: "masc user"})
	if err != nil {
		t.Fatalf("RenderComponentInto error: %v", err)
	}
	got := bodyEl.InnerHTML()
	// Must contain the input with correct attributes and greeting
	if !strings.Contains(got, `value="masc user"`) {
		t.Errorf("input value not set; got %q", got)
	}
	if !strings.Contains(got, `type="text"`) {
		t.Errorf("input type not set; got %q", got)
	}
	if !strings.Contains(got, "Hello masc user") {
		t.Errorf("greeting not set; got %q", got)
	}
}

// TestPageViewInput verifies typing into the input updates the DOM
func TestPageViewInput(t *testing.T) {
	win, err := html.NewWindowReader(strings.NewReader("<!DOCTYPE html><html><body></body></html>"))
	if err != nil {
		t.Fatalf("failed to create gost-dom window: %v", err)
	}
	// Render component with send callback
	bodyEl, send, err := masc.RenderComponentIntoWithSend(win, &PageView{input: ""})
	if err != nil {
		t.Fatalf("RenderComponentInto error: %v", err)
	}
	// Simulate user typing by sending InputMsg directly
	send(InputMsg{Value: "Alice"})
	// After auto re-render, greeting should reflect new input
	got := bodyEl.InnerHTML()
	if !strings.Contains(got, "Hello Alice") {
		t.Errorf("after send, greeting not updated; got %q", got)
	}
}

// TestPageViewClick verifies click events update the DOM under native builds.
func TestPageViewClick(t *testing.T) {
	win, err := html.NewWindowReader(strings.NewReader("<!DOCTYPE html><html><body></body></html>"))
	if err != nil {
		t.Fatalf("failed to create gost-dom window: %v", err)
	}
	bodyEl, err := masc.RenderComponentInto(win, &PageView{input: "masc user"})
	if err != nil {
		t.Fatalf("RenderComponentInto error: %v", err)
	}
	// Simulate a click on the body element (auto re-render triggers)
	bodyEl.Click()
	got := bodyEl.InnerHTML()
	// Must reflect one click in the greeting
	if !strings.Contains(got, "masc user!") {
		t.Errorf("after click, greeting not updated; got %q", got)
	}
}
