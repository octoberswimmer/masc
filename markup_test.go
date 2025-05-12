package masc

import "testing"

// TODO(slimsag): tests for other Markup

func TestNamespace(t *testing.T) {
	want := "b"
	h := Tag("a", Markup(Namespace(want)))
	if h.namespace != want {
		t.Fatalf("got namespace %q want %q", h.namespace, want)
	}
}

// TestScrollIntoView ensures the ScrollIntoView markup sets the scroll flag on the element.
func TestScrollIntoView(t *testing.T) {
	h := Tag("a", Markup(ScrollIntoView()))
	if !h.scrollIntoView {
		t.Fatal("expected scrollIntoView to be true")
	}
}
