package masc

import (
	"fmt"
	"testing"
)

type testCore struct{ Core }

func (testCore) Render(send func(Msg)) ComponentOrHTML { return Tag("p") }

type testCorePtr struct{ *Core }

func (testCorePtr) Render(send func(Msg)) ComponentOrHTML { return Tag("p") }

func TestCore(t *testing.T) {
	// Test that a standard *MyComponent with embedded Core works as we expect.
	t.Run("comp_ptr_and_core", func(t *testing.T) {
		v1 := Tag("v1")
		valid := Component(&testCore{})
		valid.Context().prevRender = v1
		if valid.Context().prevRender != v1 {
			t.Fatal("valid.Context().prevRender != v1")
		}
	})

	// Test that a non-pointer MyComponent with embedded Core does not satisfy
	// the Component interface:
	//
	//  testCore does not implement Component (Context method has pointer receiver)
	//
	t.Run("comp_and_core", func(t *testing.T) {
		isComponent := func(x interface{}) bool {
			_, ok := x.(Component)
			return ok
		}
		if isComponent(testCore{}) {
			t.Fatal("expected !isComponent(testCompCore{})")
		}
	})

	// Test what happens when a user accidentally embeds *Core instead of Core in
	// their component.
	t.Run("comp_ptr_and_core_ptr", func(t *testing.T) {
		v1 := Tag("v1")
		invalid := Component(&testCorePtr{})
		got := recoverStr(func() {
			invalid.Context().prevRender = v1
		})
		// TODO(slimsag): This would happen in user-facing code too. We should
		// create a helper for when we access a component's context, which
		// would panic with a more helpful message.
		want := "runtime error: invalid memory address or nil pointer dereference"
		if got != want {
			t.Fatalf("got panic %q want %q", got, want)
		}
	})
	t.Run("comp_and_core_ptr", func(t *testing.T) {
		v1 := Tag("v1")
		invalid := Component(testCorePtr{})
		got := recoverStr(func() {
			invalid.Context().prevRender = v1
		})
		// TODO(slimsag): This would happen in user-facing code too. We should
		// create a helper for when we access a component's context, which
		// would panic with a more helpful message.
		want := "runtime error: invalid memory address or nil pointer dereference"
		if got != want {
			t.Fatalf("got panic %q want %q", got, want)
		}
	})
}

// TODO(slimsag): TestUnmounter; Unmounter.Unmount

func TestHTML_Node(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	x := undefined()
	h := &HTML{node: x}
	if !h.Node().Equal(x.j) {
		t.Fatal("h.Node() != x")
	}
}

func send(Msg) {}

// TestHTML_reconcile_std tests that (*HTML).reconcile against an old HTML instance
// works as expected (i.e. that it updates nodes correctly).
func TestHTML_reconcile_std(t *testing.T) {
	t.Run("text_identical", func(t *testing.T) {
		ts := testSuite(t)
		defer ts.done()

		init := Text("foobar")
		init.reconcile(nil, send)

		target := Text("foobar")
		target.reconcile(init, send)
	})
	t.Run("text_diff", func(t *testing.T) {
		ts := testSuite(t)
		defer ts.done()

		init := Text("bar")
		init.reconcile(nil, send)

		target := Text("foo")
		target.reconcile(init, send)
	})
	t.Run("properties", func(t *testing.T) {
		cases := []struct {
			name        string
			initHTML    *HTML
			targetHTML  *HTML
			sortedLines [][2]int
		}{
			{
				name:        "diff",
				initHTML:    Tag("div", Markup(Property("a", 1), Property("b", "2foobar"))),
				targetHTML:  Tag("div", Markup(Property("a", 3), Property("b", "4foobar"))),
				sortedLines: [][2]int{{3, 4}, {12, 13}},
			},
			{
				name:        "remove",
				initHTML:    Tag("div", Markup(Property("a", 1), Property("b", "2foobar"))),
				targetHTML:  Tag("div", Markup(Property("a", 3))),
				sortedLines: [][2]int{{3, 4}},
			},
			{
				name:        "replaced_elem_diff",
				initHTML:    Tag("div", Markup(Property("a", 1), Property("b", "2foobar"))),
				targetHTML:  Tag("span", Markup(Property("a", 3), Property("b", "4foobar"))),
				sortedLines: [][2]int{{3, 4}, {11, 12}},
			},
			{
				name:        "replaced_elem_shared",
				initHTML:    Tag("div", Markup(Property("a", 1), Property("b", "2foobar"))),
				targetHTML:  Tag("span", Markup(Property("a", 1), Property("b", "4foobar"))),
				sortedLines: [][2]int{{3, 4}, {11, 12}},
			},
		}
		for _, tst := range cases {
			t.Run(tst.name, func(t *testing.T) {
				ts := testSuite(t)
				defer ts.multiSortedDone(tst.sortedLines...)

				tst.initHTML.reconcile(nil, send)
				ts.record("(first reconcile done)")
				tst.targetHTML.reconcile(tst.initHTML, send)
			})
		}
	})
	t.Run("attributes", func(t *testing.T) {
		cases := []struct {
			name        string
			initHTML    *HTML
			targetHTML  *HTML
			sortedLines [][2]int
		}{
			{
				name:        "diff",
				initHTML:    Tag("div", Markup(Attribute("a", 1), Attribute("b", "2foobar"))),
				targetHTML:  Tag("div", Markup(Attribute("a", 3), Attribute("b", "4foobar"))),
				sortedLines: [][2]int{{3, 4}, {12, 13}},
			},
			{
				name:        "remove",
				initHTML:    Tag("div", Markup(Attribute("a", 1), Attribute("b", "2foobar"))),
				targetHTML:  Tag("div", Markup(Attribute("a", 3))),
				sortedLines: [][2]int{{3, 4}},
			},
		}
		for _, tst := range cases {
			t.Run(tst.name, func(t *testing.T) {
				ts := testSuite(t)
				defer ts.multiSortedDone(tst.sortedLines...)

				tst.initHTML.reconcile(nil, send)
				ts.record("(first reconcile done)")
				tst.targetHTML.reconcile(tst.initHTML, send)
			})
		}
	})
	t.Run("class", func(t *testing.T) {
		cases := []struct {
			name        string
			initHTML    *HTML
			targetHTML  *HTML
			sortedLines [][2]int
		}{
			{
				name:        "multi",
				initHTML:    Tag("div", Markup(Class("a"), Class("b"))),
				targetHTML:  Tag("div", Markup(Class("a"), Class("c"))),
				sortedLines: [][2]int{{4, 5}},
			},
			{
				name:        "diff",
				initHTML:    Tag("div", Markup(Class("a", "b"))),
				targetHTML:  Tag("div", Markup(Class("a", "c"))),
				sortedLines: [][2]int{{4, 5}},
			},
			{
				name:        "remove",
				initHTML:    Tag("div", Markup(Class("a", "b"))),
				targetHTML:  Tag("div", Markup(Class("a"))),
				sortedLines: [][2]int{{4, 5}},
			},
			{
				name:        "map",
				initHTML:    Tag("div", Markup(ClassMap{"a": true, "b": true})),
				targetHTML:  Tag("div", Markup(ClassMap{"a": true})),
				sortedLines: [][2]int{{4, 5}},
			},
			{
				name:        "map_toggle",
				initHTML:    Tag("div", Markup(ClassMap{"a": true, "b": true})),
				targetHTML:  Tag("div", Markup(ClassMap{"a": true, "b": false})),
				sortedLines: [][2]int{{4, 5}},
			},
			{
				name:        "combo",
				initHTML:    Tag("div", Markup(ClassMap{"a": true, "b": true}, Class("c"))),
				targetHTML:  Tag("div", Markup(ClassMap{"a": true, "b": false}, Class("d"))),
				sortedLines: [][2]int{{4, 6}, {11, 12}},
			},
		}
		for _, tst := range cases {
			t.Run(tst.name, func(t *testing.T) {
				ts := testSuite(t)
				defer ts.multiSortedDone(tst.sortedLines...)

				tst.initHTML.reconcile(nil, send)
				ts.record("(first reconcile done)")
				tst.targetHTML.reconcile(tst.initHTML, send)
			})
		}
	})
	t.Run("dataset", func(t *testing.T) {
		cases := []struct {
			name        string
			initHTML    *HTML
			targetHTML  *HTML
			sortedLines [][2]int
		}{
			{
				name:        "diff",
				initHTML:    Tag("div", Markup(Data("a", "1"), Data("b", "2foobar"))),
				targetHTML:  Tag("div", Markup(Data("a", "3"), Data("b", "4foobar"))),
				sortedLines: [][2]int{{5, 6}, {14, 15}},
			},
			{
				name:        "remove",
				initHTML:    Tag("div", Markup(Data("a", "1"), Data("b", "2foobar"))),
				targetHTML:  Tag("div", Markup(Data("a", "3"))),
				sortedLines: [][2]int{{5, 6}},
			},
		}
		for _, tst := range cases {
			t.Run(tst.name, func(t *testing.T) {
				ts := testSuite(t)
				defer ts.multiSortedDone(tst.sortedLines...)

				tst.initHTML.reconcile(nil, send)
				ts.record("(first reconcile done)")
				tst.targetHTML.reconcile(tst.initHTML, send)
			})
		}
	})
	t.Run("style", func(t *testing.T) {
		cases := []struct {
			name        string
			initHTML    *HTML
			targetHTML  *HTML
			sortedLines [][2]int
		}{
			{
				name:        "diff",
				initHTML:    Tag("div", Markup(Style("a", "1"), Style("b", "2foobar"))),
				targetHTML:  Tag("div", Markup(Style("a", "3"), Style("b", "4foobar"))),
				sortedLines: [][2]int{{6, 7}, {15, 16}},
			},
			{
				name:        "remove",
				initHTML:    Tag("div", Markup(Style("a", "1"), Style("b", "2foobar"))),
				targetHTML:  Tag("div", Markup(Style("a", "3"))),
				sortedLines: [][2]int{{6, 7}},
			},
		}
		for _, tst := range cases {
			t.Run(tst.name, func(t *testing.T) {
				ts := testSuite(t)
				defer ts.multiSortedDone(tst.sortedLines...)

				tst.initHTML.reconcile(nil, send)
				ts.record("(first reconcile done)")
				tst.targetHTML.reconcile(tst.initHTML, send)
			})
		}
	})
	t.Run("event_listener", func(t *testing.T) {
		// TODO(pdf): Mock listener functions for equality testing
		ts := testSuite(t)
		defer ts.done()

		initEventListeners := []Applyer{
			&EventListener{Name: "click"},
			&EventListener{Name: "keydown"},
		}
		prev := Tag("div", Markup(initEventListeners...))
		prev.reconcile(nil, send)
		ts.record("(expected two added event listeners above)")
		for i, m := range initEventListeners {
			listener := m.(*EventListener)
			if listener.wrapper == nil {
				t.Fatalf("listener %d wrapper == nil: %+v", i, listener)
			}
		}

		targetEventListeners := []Applyer{
			&EventListener{Name: "click"},
		}
		h := Tag("div", Markup(targetEventListeners...))
		h.reconcile(prev, send)
		ts.record("(expected two removed, one added event listeners above)")
		for i, m := range targetEventListeners {
			listener := m.(*EventListener)
			if listener.wrapper == nil {
				t.Fatalf("listener %d wrapper == nil: %+v", i, listener)
			}
		}
	})

	// TODO(pdf): test (*HTML).reconcile child mutations, and value/checked properties
	// TODO(pdf): test multi-pass reconcile of persistent component pointer children, ref: https://github.com/hexops/vecty/pull/124
}

// TestHTML_reconcile_nil tests that (*HTML).reconcile(nil) works as expected (i.e.
// that it creates nodes correctly).
func TestHTML_reconcile_nil(t *testing.T) {
	t.Run("one_of_tag_or_text", func(t *testing.T) {
		got := recoverStr(func() {
			h := &HTML{text: "hello", tag: "div"}
			h.reconcile(nil, send)
		})
		want := "masc: internal error (only one of HTML.tag or HTML.text may be set)"
		if got != want {
			t.Fatalf("got panic %q want %q", got, want)
		}
	})
	t.Run("unsafe_text", func(t *testing.T) {
		got := recoverStr(func() {
			h := &HTML{text: "hello", innerHTML: "foobar"}
			h.reconcile(nil, send)
		})
		want := "masc: only HTML may have UnsafeHTML attribute"
		if got != want {
			t.Fatalf("got panic %q want %q", got, want)
		}
	})
	t.Run("create_element", func(t *testing.T) {
		ts := testSuite(t)
		defer ts.done()

		h := Tag("strong")
		h.reconcile(nil, send)
	})
	t.Run("create_element_ns", func(t *testing.T) {
		ts := testSuite(t)
		defer ts.done()

		h := Tag("strong", Markup(Namespace("foobar")))
		h.reconcile(nil, send)
	})
	t.Run("create_text_node", func(t *testing.T) {
		ts := testSuite(t)
		defer ts.done()

		h := Text("hello")
		h.reconcile(nil, send)
	})
	t.Run("inner_html", func(t *testing.T) {
		ts := testSuite(t)
		defer ts.done()

		h := Tag("div", Markup(UnsafeHTML("<p>hello</p>")))
		h.reconcile(nil, send)
	})
	t.Run("properties", func(t *testing.T) {
		ts := testSuite(t)
		defer ts.sortedDone(3, 4)

		h := Tag("div", Markup(Property("a", 1), Property("b", "2foobar")))
		h.reconcile(nil, send)
	})
	t.Run("attributes", func(t *testing.T) {
		ts := testSuite(t)
		defer ts.sortedDone(3, 4)

		h := Tag("div", Markup(Attribute("a", 1), Attribute("b", "2foobar")))
		h.reconcile(nil, send)
	})
	t.Run("dataset", func(t *testing.T) {
		ts := testSuite(t)
		defer ts.sortedDone(5, 6)

		h := Tag("div", Markup(Data("a", "1"), Data("b", "2foobar")))
		h.reconcile(nil, send)
	})
	t.Run("style", func(t *testing.T) {
		ts := testSuite(t)
		defer ts.sortedDone(6, 7)

		h := Tag("div", Markup(Style("a", "1"), Style("b", "2foobar")))
		h.reconcile(nil, send)
	})
	t.Run("add_event_listener", func(t *testing.T) {
		ts := testSuite(t)
		defer ts.done()

		e0 := &EventListener{Name: "click"}
		e1 := &EventListener{Name: "keydown"}
		h := Tag("div", Markup(e0, e1))
		h.reconcile(nil, send)
		if e0.wrapper == nil {
			t.Fatal("e0.wrapper == nil")
		}
		if e1.wrapper == nil {
			t.Fatal("e1.wrapper == nil")
		}
	})
	t.Run("children", func(t *testing.T) {
		ts := testSuite(t)
		defer ts.done()

		var compRenderCalls int
		compRender := Tag("div")
		comp := &componentFunc{
			id: "foobar",
			render: func() ComponentOrHTML {
				compRenderCalls++
				return compRender
			},
		}
		h := Tag("div", Tag("div", comp))
		h.reconcile(nil, send)
		if compRenderCalls != 1 {
			t.Fatal("compRenderCalls != 1")
		}
		if comp.Context().prevRenderComponent.(*componentFunc).id != comp.id {
			t.Fatal("comp.Context().prevRenderComponent.(*componentFunc).id != comp.id")
		}
		if comp.Context().prevRender != compRender {
			t.Fatal("comp.Context().prevRender != compRender")
		}
	})
	t.Run("children_render_nil", func(t *testing.T) {
		ts := testSuite(t)
		defer ts.done()

		var compRenderCalls int
		comp := &componentFunc{
			id: "foobar",
			render: func() ComponentOrHTML {
				compRenderCalls++
				return nil
			},
		}
		h := Tag("div", Tag("div", comp))
		h.reconcile(nil, send)
		if compRenderCalls != 1 {
			t.Fatal("compRenderCalls != 1")
		}
		if comp.Context().prevRenderComponent.(*componentFunc).id != comp.id {
			t.Fatal("comp.Context().prevRenderComponent.(*componentFunc).id != comp.id")
		}
		if comp.Context().prevRender == nil {
			t.Fatal("comp.Context().prevRender == nil")
		}
	})
}

func TestTag(t *testing.T) {
	markupCalled := false
	want := "foobar"
	h := Tag(want, Markup(markupFunc(func(h *HTML) {
		markupCalled = true
	})))
	if !markupCalled {
		t.Fatal("expected markup to be applied")
	}
	if h.tag != want {
		t.Fatalf("got tag %q want tag %q", h.text, want)
	}
	if h.text != "" {
		t.Fatal("expected no text")
	}
}

func TestText(t *testing.T) {
	markupCalled := false
	want := "Hello world!"
	h := Text(want, Markup(markupFunc(func(h *HTML) {
		markupCalled = true
	})))
	if !markupCalled {
		t.Fatal("expected markup to be applied")
	}
	if h.text != want {
		t.Fatalf("got text %q want text %q", h.text, want)
	}
	if h.tag != "" {
		t.Fatal("expected no tag")
	}
}

// TestRerender_nil tests that Rerender panics when the component argument is
// nil.
func TestRerender_nil(t *testing.T) {
	gotPanic := ""
	func() {
		defer func() {
			r := recover()
			if r != nil {
				gotPanic = fmt.Sprint(r)
			}
		}()
		rerender(nil, send)
	}()
	expected := "masc: Rerender illegally called with a nil Component argument"
	if gotPanic != expected {
		t.Fatalf("got panic %q expected %q", gotPanic, expected)
	}
}

// TestRerender_no_prevRender tests the behavior of Rerender when there is no
// previous render.
func TestRerender_no_prevRender(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	got := recoverStr(func() {
		rerender(&componentFunc{
			render: func() ComponentOrHTML {
				panic("expected no Render call")
			},
			skipRender: func(prev Component) bool {
				panic("expected no SkipRender call")
			},
		}, send)
	})
	want := "masc: Rerender invoked on Component that has never been rendered"
	if got != want {
		t.Fatalf("got panic %q expected %q", got, want)
	}
}

// TestRerender_identical tests the behavior of Rerender when there is a
// previous render which is identical to the new render.
func TestRerender_identical(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
	ts.strings.mock(`global.Get("document").Get("readyState")`, "complete")
	ts.strings.mock(`global.Get("document").Call("querySelector", "body").Get("nodeName")`, "BODY")
	ts.truthies.mock(`global.Get("document").Call("querySelector", "body")`, true)

	// Perform the initial render of the component.
	render := Tag("body")
	var renderCalled, skipRenderCalled int
	comp := &componentFunc{
		id: "original",
		render: func() ComponentOrHTML {
			renderCalled++
			return render
		},
	}
	RenderBody(comp, send)
	if renderCalled != 1 {
		t.Fatal("renderCalled != 1")
	}
	if comp.Context().prevRender != render {
		t.Fatal("comp.Context().prevRender != render")
	}
	if comp.Context().prevRenderComponent.(*componentFunc).id != "original" {
		t.Fatal(`comp.Context().prevRenderComponent.(*componentFunc).id != "original"`)
	}

	// Perform a re-render.
	newRender := Tag("body")
	comp.id = "modified"
	comp.render = func() ComponentOrHTML {
		renderCalled++
		return newRender
	}
	comp.skipRender = func(prev Component) bool {
		if comp.id != "modified" {
			panic(`comp.id != "modified"`)
		}
		if comp.Context().prevRenderComponent.(*componentFunc).id != "original" {
			panic(`comp.Context().prevRenderComponent.(*componentFunc).id != "original"`)
		}
		if prev.(*componentFunc).id != "original" {
			panic(`prev.(*componentFunc).id != "original"`)
		}
		skipRenderCalled++
		return false
	}
	rerender(comp, send)

	// Invoke the render callback.
	ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
	ts.invokeCallbackRequestAnimationFrame(0)

	if renderCalled != 2 {
		t.Fatal("renderCalled != 2")
	}
	if skipRenderCalled != 1 {
		t.Fatal("skipRenderCalled != 1")
	}
	if comp.Context().prevRender != newRender {
		t.Fatal("comp.Context().prevRender != newRender")
	}
	if comp.Context().prevRenderComponent.(*componentFunc).id != "modified" {
		t.Fatal(`comp.Context().prevRenderComponent.(*componentFunc).id != "modified"`)
	}
}

// TestRerender_change tests the behavior of Rerender when there is a
// previous render which is different from the new render.
func TestRerender_change(t *testing.T) {
	cases := []struct {
		name      string
		newRender *HTML
	}{
		{
			name:      "new_child",
			newRender: Tag("body", Tag("div")),
		},
		// TODO(slimsag): bug! nil produces <noscript> and we incorrectly try
		// to replace <body> with it! We should panic & warn the user.
		//{
		//	name:      "nil",
		//	newRender: nil,
		//},
	}
	for _, tst := range cases {
		t.Run(tst.name, func(t *testing.T) {
			ts := testSuite(t)
			defer ts.done()

			ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
			ts.strings.mock(`global.Get("document").Get("readyState")`, "complete")
			ts.strings.mock(`global.Get("document").Call("querySelector", "body").Get("nodeName")`, "BODY")
			ts.truthies.mock(`global.Get("document").Call("querySelector", "body")`, true)

			// Perform the initial render of the component.
			render := Tag("body")
			var renderCalled, skipRenderCalled int
			comp := &componentFunc{
				id: "original",
				render: func() ComponentOrHTML {
					renderCalled++
					return render
				},
			}
			RenderBody(comp, send)
			ts.record("(expect body to be set now)")
			if renderCalled != 1 {
				t.Fatal("renderCalled != 1")
			}
			if comp.Context().prevRender != render {
				t.Fatal("comp.Context().prevRender != render")
			}
			if comp.Context().prevRenderComponent.(*componentFunc).id != "original" {
				t.Fatal(`comp.Context().prevRenderComponent.(*componentFunc).id != "original"`)
			}

			// Perform a re-render.
			comp.id = "modified"
			comp.render = func() ComponentOrHTML {
				renderCalled++
				return tst.newRender
			}
			comp.skipRender = func(prev Component) bool {
				if comp.id != "modified" {
					panic(`comp.id != "modified"`)
				}
				if comp.Context().prevRenderComponent.(*componentFunc).id != "original" {
					panic(`comp.Context().prevRenderComponent.(*componentFunc).id != "original"`)
				}
				if prev.(*componentFunc).id != "original" {
					panic(`prev.(*componentFunc).id != "original"`)
				}
				skipRenderCalled++
				return false
			}
			rerender(comp, send)

			// Invoke the render callback.
			ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
			ts.invokeCallbackRequestAnimationFrame(0)

			if renderCalled != 2 {
				t.Fatal("renderCalled != 2")
			}
			if skipRenderCalled != 1 {
				t.Fatal("skipRenderCalled != 1")
			}
			if comp.Context().prevRender != tst.newRender {
				t.Fatal("comp.Context().prevRender != tst.newRender")
			}
			if comp.Context().prevRenderComponent.(*componentFunc).id != "modified" {
				t.Fatal(`comp.Context().prevRenderComponent.(*componentFunc).id != "modified"`)
			}
		})
	}
}

// TestRerender_Nested tests the behavior of Rerender when there is a
// nested Component that is exchanged for *HTML.
func TestRerender_Nested(t *testing.T) {
	cases := []struct {
		name                     string
		initialRender, newRender ComponentOrHTML
	}{
		{
			name:          "html_to_component",
			initialRender: Tag("body"),
			newRender: &componentFunc{
				render: func() ComponentOrHTML {
					return Tag("body", Tag("div"))
				},
				skipRender: func(Component) bool {
					return false
				},
			},
		},
		{
			name: "component_to_html",
			initialRender: &componentFunc{
				render: func() ComponentOrHTML {
					return Tag("body")
				},
				skipRender: func(Component) bool {
					return false
				},
			},
			newRender: Tag("body", Tag("div")),
		},
	}
	for _, tst := range cases {
		t.Run(tst.name, func(t *testing.T) {
			ts := testSuite(t)
			defer ts.done()

			ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
			ts.strings.mock(`global.Get("document").Get("readyState")`, "complete")
			ts.strings.mock(`global.Get("document").Call("querySelector", "body").Get("nodeName")`, "BODY")
			ts.truthies.mock(`global.Get("document").Call("querySelector", "body")`, true)

			// Perform the initial render of the component.
			var renderCalled, skipRenderCalled int
			comp := &componentFunc{
				id: "original",
				render: func() ComponentOrHTML {
					renderCalled++
					return tst.initialRender
				},
			}
			RenderBody(comp, send)
			ts.record("(expect body to be set now)")
			if renderCalled != 1 {
				t.Fatal("renderCalled != 1")
			}
			if comp.Context().prevRender != tst.initialRender {
				t.Fatal("comp.Context().prevRender != render")
			}
			if comp.Context().prevRenderComponent.(*componentFunc).id != "original" {
				t.Fatal(`comp.Context().prevRenderComponent.(*componentFunc).id != "original"`)
			}

			// Perform a re-render.
			comp.id = "modified"
			comp.render = func() ComponentOrHTML {
				renderCalled++
				return tst.newRender
			}
			comp.skipRender = func(prev Component) bool {
				if comp.id != "modified" {
					panic(`comp.id != "modified"`)
				}
				if comp.Context().prevRenderComponent.(*componentFunc).id != "original" {
					panic(`comp.Context().prevRenderComponent.(*componentFunc).id != "original"`)
				}
				if prev.(*componentFunc).id != "original" {
					panic(`prev.(*componentFunc).id != "original"`)
				}
				skipRenderCalled++
				return false
			}
			rerender(comp, send)

			// Invoke the render callback.
			ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
			ts.invokeCallbackRequestAnimationFrame(0)

			if skipRenderCalled != 1 {
				t.Fatal("skipRenderCalled != 1")
			}
			if comp.Context().prevRender != tst.newRender {
				t.Fatal("comp.Context().prevRender != tst.newRender")
			}
			if comp.Context().prevRenderComponent.(*componentFunc).id != "modified" {
				t.Fatal(`comp.Context().prevRenderComponent.(*componentFunc).id != "modified"`)
			}
		})
	}
}

type persistentComponentBody struct {
	Core
}

func (c *persistentComponentBody) Render(send func(Msg)) ComponentOrHTML {
	return Tag(
		"body",
		&persistentComponent{},
	)
}

var (
	lastRenderedComponent Component
	renderCount           int
)

type persistentComponent struct {
	Core
}

func (c *persistentComponent) Render(send func(Msg)) ComponentOrHTML {
	if lastRenderedComponent == nil {
		lastRenderedComponent = c
	} else if lastRenderedComponent != c {
		panic("unexpected last rendered component")
	}
	renderCount++
	return Tag("div")
}

// TestRerender_persistent tests the behavior of rendering persistent
// components.
func TestRerender_persistent(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
	ts.strings.mock(`global.Get("document").Get("readyState")`, "complete")
	ts.strings.mock(`global.Get("document").Call("querySelector", "body").Get("nodeName")`, "BODY")
	ts.truthies.mock(`global.Get("document").Call("querySelector", "body")`, true)

	lastRenderedComponent = nil
	renderCount = 0

	comp := &persistentComponentBody{}
	// Perform the initial render of the component.
	RenderBody(comp, send)

	if renderCount != 1 {
		t.Fatal("renderCount != 1")
	}

	// Perform a re-render.
	rerender(comp, send)

	// Invoke the render callback.
	ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
	ts.invokeCallbackRequestAnimationFrame(0)

	if renderCount != 2 {
		t.Fatal("renderCount != 2")
	}

	// Perform a re-render.
	rerender(comp, send)

	// Invoke the render callback.
	ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
	ts.invokeCallbackRequestAnimationFrame(0)

	if renderCount != 3 {
		t.Fatal("renderCount != 3")
	}
}

type persistentComponentBody2 struct {
	Core
}

func (c *persistentComponentBody2) Render(send func(Msg)) ComponentOrHTML {
	return Tag(
		"body",
		&persistentWrapperComponent{},
	)
}

type persistentWrapperComponent struct {
	Core
}

func (c *persistentWrapperComponent) Render(send func(Msg)) ComponentOrHTML {
	return &persistentComponent{}
}

// TestRerender_persistent_direct tests the behavior of rendering persistent
// components that are directly returned by Render().
func TestRerender_persistent_direct(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
	ts.strings.mock(`global.Get("document").Get("readyState")`, "complete")
	ts.strings.mock(`global.Get("document").Call("querySelector", "body").Get("nodeName")`, "BODY")
	ts.truthies.mock(`global.Get("document").Call("querySelector", "body")`, true)

	lastRenderedComponent = nil
	renderCount = 0

	comp := &persistentComponentBody2{}
	// Perform the initial render of the component.
	RenderBody(comp, send)

	if renderCount != 1 {
		t.Fatal("renderCount != 1")
	}

	// Perform a re-render.
	rerender(comp, send)

	// Invoke the render callback.
	ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
	ts.invokeCallbackRequestAnimationFrame(0)

	if renderCount != 2 {
		t.Fatal("renderCount != 2")
	}

	// Perform a re-render.
	rerender(comp, send)

	// Invoke the render callback.
	ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
	ts.invokeCallbackRequestAnimationFrame(0)

	if renderCount != 3 {
		t.Fatal("renderCount != 3")
	}
}

// TestRenderBody_ExpectsBody tests that RenderBody panics when something other
// than a "body" tag is rendered by the component.
func TestRenderBody_ExpectsBody(t *testing.T) {
	cases := []struct {
		name      string
		render    *HTML
		wantPanic string
	}{
		{
			name:      "text",
			render:    Text("Hello world!"),
			wantPanic: "masc: RenderBody: expected Component.Render to return a \"body\", found \"\"", // TODO(slimsag): error message bug
		},
		{
			name:      "div",
			render:    Tag("div"),
			wantPanic: "masc: RenderBody: expected Component.Render to return a \"body\", found \"div\"",
		},
		{
			name:      "nil",
			render:    nil,
			wantPanic: "masc: RenderBody: expected Component.Render to return a \"body\", found \"noscript\"",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ts := testSuite(t)
			defer ts.done()

			ts.strings.mock(`global.Get("document").Call("querySelector", "body").Get("nodeName")`, "BODY")
			ts.truthies.mock(`global.Get("document").Call("querySelector", "body")`, true)

			var gotPanic string
			func() {
				defer func() {
					r := recover()
					if r != nil {
						gotPanic = fmt.Sprint(r)
					}
				}()
				RenderBody(&componentFunc{
					render: func() ComponentOrHTML {
						return c.render
					},
					skipRender: func(prev Component) bool { return false },
				}, send)
			}()
			if c.wantPanic != gotPanic {
				t.Fatalf("want panic:\n%q\ngot panic:\n%q", c.wantPanic, gotPanic)
			}
		})
	}
}

// TestRenderBody_RenderSkipper_Skip tests that RenderBody panics when the
// component's SkipRender method returns skip == true.
func TestRenderBody_RenderSkipper_Skip(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	ts.truthies.mock(`global.Get("document").Call("querySelector", "body")`, true)

	comp := &componentFunc{
		render: func() ComponentOrHTML {
			return Tag("body")
		},
		skipRender: func(prev Component) bool {
			return true
		},
	}
	fakePrevRender := *comp
	comp.Context().prevRenderComponent = &fakePrevRender
	got := recoverStr(func() {
		RenderBody(comp, send)
	})
	want := "masc: RenderBody: Component.SkipRender illegally returned true"
	if got != want {
		t.Fatalf("got panic %q want %q", got, want)
	}
}

// TestRenderBody_Standard_loaded tests that RenderBody properly handles the
// standard case of rendering into the "body" tag when the DOM is in a loaded
// state.
func TestRenderBody_Standard_loaded(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	ts.strings.mock(`global.Get("document").Get("readyState")`, "loaded")
	ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
	ts.strings.mock(`global.Get("document").Call("querySelector", "body").Get("nodeName")`, "BODY")
	ts.truthies.mock(`global.Get("document").Call("querySelector", "body")`, true)

	RenderBody(&componentFunc{
		render: func() ComponentOrHTML {
			return Tag("body")
		},
	}, send)
}

// TestRenderBody_Standard_loading tests that RenderBody properly handles the
// standard case of rendering into the "body" tag when the DOM is in a loading
// state.
func TestRenderBody_Standard_loading(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	ts.strings.mock(`global.Get("document").Get("readyState")`, "loading")
	ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
	ts.strings.mock(`global.Get("document").Call("querySelector", "body").Get("nodeName")`, "BODY")
	ts.truthies.mock(`global.Get("document").Call("querySelector", "body")`, true)

	RenderBody(&componentFunc{
		render: func() ComponentOrHTML {
			return Tag("body")
		},
	}, send)

	ts.record("(invoking DOMContentLoaded event listener)")
	ts.invokeCallbackDOMContentLoaded()
}

// TestRenderBody_Nested tests that RenderBody properly handles nested
// Components.
func TestRenderBody_Nested(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	ts.strings.mock(`global.Get("document").Get("readyState")`, "complete")
	ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
	ts.strings.mock(`global.Get("document").Call("querySelector", "body").Get("nodeName")`, "BODY")
	ts.truthies.mock(`global.Get("document").Call("querySelector", "body")`, true)

	RenderBody(&componentFunc{
		render: func() ComponentOrHTML {
			return &componentFunc{
				render: func() ComponentOrHTML {
					return &componentFunc{
						render: func() ComponentOrHTML {
							return Tag("body")
						},
					}
				},
			}
		},
	}, send)
}

// TestSetTitle tests that the SetTitle function performs the correct DOM
// operations.
func TestSetTitle(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	SetTitle("foobartitle")
}

// TestAddStylesheet tests that the AddStylesheet performs the correct DOM
// operations.
func TestAddStylesheet(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	AddStylesheet("https://google.com/foobar.css")
}

func TestKeyedChild_DifferentType(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
	ts.strings.mock(`global.Get("document").Get("readyState")`, "complete")
	ts.strings.mock(`global.Get("document").Call("querySelector", "body").Get("nodeName")`, "BODY")
	ts.truthies.mock(`global.Get("document").Call("querySelector", "body")`, true)

	comp := &componentFunc{
		render: func() ComponentOrHTML {
			return Tag(
				"body",
				Tag(
					"tag1",
					Markup(ElementKey("key")),
				),
			)
		},
		skipRender: func(prev Component) bool { return false },
	}

	RenderBody(comp, send)

	rerender := func() {
		rerender(comp, send)
		ts.isUndefined.mock(`global.Call("requestAnimationFrame", func)`, 0)
		ts.invokeCallbackRequestAnimationFrame(0)
	}

	// Re-render
	rerender()

	comp.render = func() ComponentOrHTML {
		return Tag(
			"body",
			Tag(
				"tag2",
				Markup(ElementKey("key")),
			),
		)
	}

	rerender()
}

// TestEventListenerRelease_TagChange tests that event listener wrappers are
// released when a node's tag changes during reconciliation.
func TestEventListenerRelease_TagChange(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	// Create initial element with event listeners
	e0 := &EventListener{Name: "click"}
	e1 := &EventListener{Name: "keydown"}
	prev := Tag("div", Markup(e0, e1))
	prev.reconcile(nil, send)

	// Verify wrappers were created
	if e0.wrapper == nil {
		t.Fatal("e0.wrapper == nil after initial reconcile")
	}
	if e1.wrapper == nil {
		t.Fatal("e1.wrapper == nil after initial reconcile")
	}

	// Get the wrapper implementations to check released state
	w0 := e0.wrapper.(*jsFuncImpl)
	w1 := e1.wrapper.(*jsFuncImpl)

	if w0.released {
		t.Fatal("w0 should not be released before tag change")
	}
	if w1.released {
		t.Fatal("w1 should not be released before tag change")
	}

	// Reconcile with a different tag - this should release old event listeners
	next := Tag("span") // Different tag, no event listeners
	next.reconcile(prev, send)

	// Verify old wrappers were released
	if !w0.released {
		t.Fatal("w0 should be released after tag change")
	}
	if !w1.released {
		t.Fatal("w1 should be released after tag change")
	}
}

// TestEventListenerRelease_Unmount tests that event listener wrappers are
// released when a node is unmounted.
func TestEventListenerRelease_Unmount(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	// Create element with event listeners
	e0 := &EventListener{Name: "click"}
	h := Tag("div", Markup(e0))
	h.reconcile(nil, send)

	// Verify wrapper was created
	if e0.wrapper == nil {
		t.Fatal("e0.wrapper == nil after reconcile")
	}

	w0 := e0.wrapper.(*jsFuncImpl)
	if w0.released {
		t.Fatal("w0 should not be released before unmount")
	}

	// Unmount the element
	unmount(h)

	// Verify wrapper was released
	if !w0.released {
		t.Fatal("w0 should be released after unmount")
	}
}

// TestEventListenerRelease_NestedUnmount tests that event listener wrappers
// are released for nested elements when parent is unmounted.
func TestEventListenerRelease_NestedUnmount(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	// Create nested elements with event listeners
	childListener := &EventListener{Name: "click"}
	child := Tag("button", Markup(childListener))
	parent := Tag("div", child)
	parent.reconcile(nil, send)

	// Verify wrapper was created
	if childListener.wrapper == nil {
		t.Fatal("childListener.wrapper == nil after reconcile")
	}

	w := childListener.wrapper.(*jsFuncImpl)
	if w.released {
		t.Fatal("wrapper should not be released before unmount")
	}

	// Unmount the parent (should also release child's listeners)
	unmount(parent)

	// Verify child's wrapper was released
	if !w.released {
		t.Fatal("child wrapper should be released after parent unmount")
	}
}

// TestCopyComponent_ClearsPrevRender tests that copyComponent clears
// prevRender and prevRenderComponent on the copy to prevent memory leaks.
func TestCopyComponent_ClearsPrevRender(t *testing.T) {
	comp := &componentFunc{
		id: "original",
		render: func() ComponentOrHTML {
			return Tag("div")
		},
	}

	// Set up prevRender and prevRenderComponent
	prevHTML := Tag("span")
	prevComp := &componentFunc{id: "prev"}
	comp.Context().prevRender = prevHTML
	comp.Context().prevRenderComponent = prevComp

	// Copy the component
	cpy := copyComponent(comp)

	// Verify the copy has cleared prevRender fields
	if cpy.Context().prevRender != nil {
		t.Fatal("copyComponent should clear prevRender on the copy")
	}
	if cpy.Context().prevRenderComponent != nil {
		t.Fatal("copyComponent should clear prevRenderComponent on the copy")
	}

	// Verify original is unchanged
	if comp.Context().prevRender != prevHTML {
		t.Fatal("copyComponent should not modify original's prevRender")
	}
	if comp.Context().prevRenderComponent != prevComp {
		t.Fatal("copyComponent should not modify original's prevRenderComponent")
	}
}

// TestCopyComponent_WithCopier_ClearsPrevRender tests that copyComponent
// clears prevRender fields even when the component implements Copier.
func TestCopyComponent_WithCopier_ClearsPrevRender(t *testing.T) {
	comp := &componentWithCopier{
		id: "original",
	}

	// Set up prevRender and prevRenderComponent
	prevHTML := Tag("span")
	prevComp := &componentFunc{id: "prev"}
	comp.Context().prevRender = prevHTML
	comp.Context().prevRenderComponent = prevComp

	// Copy the component
	cpy := copyComponent(comp)

	// Verify the copy has cleared prevRender fields
	if cpy.Context().prevRender != nil {
		t.Fatal("copyComponent should clear prevRender on the copy")
	}
	if cpy.Context().prevRenderComponent != nil {
		t.Fatal("copyComponent should clear prevRenderComponent on the copy")
	}
}

// componentWithCopier is a test component that implements the Copier interface.
type componentWithCopier struct {
	Core
	id string
}

func (c *componentWithCopier) Render(send func(Msg)) ComponentOrHTML {
	return Tag("div")
}

func (c *componentWithCopier) Copy() Component {
	cpy := *c
	return &cpy
}

// TestEventListenerWrapper_NilAfterRelease tests that the wrapper field
// is set to nil after release to help GC break reference chains.
func TestEventListenerWrapper_NilAfterRelease(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	// Create element with event listener
	e0 := &EventListener{Name: "click"}
	prev := Tag("div", Markup(e0))
	prev.reconcile(nil, send)

	// Verify wrapper was created
	if e0.wrapper == nil {
		t.Fatal("e0.wrapper == nil after reconcile")
	}

	// Reconcile with element that has no listeners (triggers removeProperties)
	next := Tag("div") // Same tag, no event listeners
	next.reconcile(prev, send)

	// Verify wrapper was released and nilled
	if e0.wrapper != nil {
		t.Fatal("e0.wrapper should be nil after release in removeProperties")
	}
}

// TestReleaseEventListeners_NilsWrapper tests that releaseEventListeners
// sets wrapper to nil after release.
func TestReleaseEventListeners_NilsWrapper(t *testing.T) {
	ts := testSuite(t)
	defer ts.done()

	// Create element with event listener
	e0 := &EventListener{Name: "click"}
	h := Tag("div", Markup(e0))
	h.reconcile(nil, send)

	// Verify wrapper was created
	if e0.wrapper == nil {
		t.Fatal("e0.wrapper == nil after reconcile")
	}

	// Directly call releaseEventListeners
	h.releaseEventListeners()

	// Verify wrapper was nilled
	if e0.wrapper != nil {
		t.Fatal("e0.wrapper should be nil after releaseEventListeners")
	}
}
