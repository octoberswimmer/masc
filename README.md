# masc

Masc combines the state management of [Bubble Tea](https://github.com/charmbracelet/bubbletea/) with the [Vecty](https://github.com/hexops/vecty) view rendering model.  The result is a library for building browser applications in Go using [The Elm Architecture](https://guide.elm-lang.org/architecture/).

Vecty components are stateless, or, at least, agnostic about how state is managed.

Bubble Tea models are stateful, or, at least, opinionated about how state should be managed.

Masc models look just like Bubble Tea models, except that they return HTML or
other components rather than strings when being rendered.  This is just like
vecty.  The vecty rendering engine is used to update the browser.

Masc components look just like Vecty components, except the Render function
takes a `func(Msg)` parameter.  This function, called `send` by convention, is
used to send messages to the running program to update its state.

Stateless components implement the Component interface, i.e. have a `Render(func(Msg) ComponentOrHTML` function.

Models are Components that also implement the Model interface, i.e. have `Init() Cmd` and `Update(Msg) (Model, Cmd)` functions.

That is, <b>m</b>odels <b>a</b>re <b>s</b>tateful <b>c</b>omponents.

## Example

Here's a basic Hello World example.

[embedmd]:# (example/hellomasc/hellomasc.go)
```go
package main

import (
	"strings"

	"github.com/octoberswimmer/masc"
	"github.com/octoberswimmer/masc/elem"
	"github.com/octoberswimmer/masc/event"
)

func main() {
	masc.SetTitle("Hello masc User!")
	// Initialize a model
	m := &PageView{}
	pgm := masc.NewProgram(m)
	_, err := pgm.Run()
	if err != nil {
		panic(err)
	}
}

type ClickMsg struct{}

// InputMsg is sent when the text input changes.
type InputMsg struct{ Value string }

// PageView is our main page component.
type PageView struct {
	masc.Core

	// The model state is the number of clicks
	clicks int
	input  string
}

func (p *PageView) Init() masc.Cmd {
	p.input = "masc user"
	return nil
}

// PageView is a masc.Model; it has Init and Update functions
func (p *PageView) Update(msg masc.Msg) (masc.Model, masc.Cmd) {
	switch m := msg.(type) {
	case ClickMsg:
		// Update the model state when we get a click message
		p.clicks++
	case InputMsg:
		// Update the model state when input changes
		p.input = m.Value
	}
	return p, nil
}

// Render implements the masc.Component interface.
func (p *PageView) Render(send func(masc.Msg)) masc.ComponentOrHTML {
	return elem.Body(
		masc.Markup(
			event.Click(func(e *masc.Event) {
				// Send a click message upon the click event
				send(ClickMsg{})
			}),
		),
		// Text input that echoes into the greeting
		elem.Input(
			masc.Markup(
				masc.Property("type", "text"),
				masc.Property("value", p.input),
				event.Input(func(e *masc.Event) {
					// Grab the new value from the event target
					v := e.Value.Get("target").Get("value").String()
					send(InputMsg{Value: v})
				}),
			),
		),
		elem.Break(),
		// Greeting text includes both input and clicks
		masc.Text("Hello "+p.input+""+strings.Repeat("!", p.clicks)),
	)
}
```

Additional examples, including a todo app,
are in the [example](example/) directory.

## Running Examples with the masc CLI

The recommended way to run masc applications is using the built-in `masc serve` command:

```bash
# Install the masc CLI
go install github.com/octoberswimmer/masc/cmd/masc@latest

# Serve the hello world example
masc serve ./example/hellomasc/

# Serve with custom port
masc serve -p 3000 ./example/hellomasc/
```

The `masc serve` command:
- Automatically builds your Go application to WebAssembly
- Serves the application on a local development server (default port 8000)
- Watches for file changes and automatically rebuilds
- Opens your default browser to the application
- Supports Go workspaces and handles module dependencies intelligently

### Alternative: Using wasmserve

Examples can also be run using [wasmserve](https://github.com/hajimehoshi/wasmserve) for manual WebAssembly builds.

## Pure-Go DOM Testing with gost-dom

To test masc components in pure Go without a browser, you can use [gost-dom/browser](https://github.com/gost-dom/browser).  Under native (non-WASM) builds, masc will render into a gost-dom Window, which you can inspect and drive directly:

```go
package yourpkg_test

import (
    "strings"
    "testing"

    "github.com/gost-dom/browser/html"
    "github.com/octoberswimmer/masc"
    "github.com/your/module/foo"
)

func TestMyWidget(t *testing.T) {
    // Create a gost-dom window and parse initial HTML
    win, err := html.NewWindowReader(
        strings.NewReader("<!DOCTYPE html><html><body></body></html>"),
    )
    if err != nil {
        t.Fatal(err)
    }

    // Render component with automatic re-render support
    body, send, err := masc.RenderComponentIntoWithSend(win, foo.NewWidget())
    if err != nil {
        t.Fatal(err)
    }

    // Simulate user interactions:
    // 1) Click a specific button:
    btn := win.Document().QuerySelector("button").(html.HTMLElement)
    btn.Click()

    // 2) Click the body element:
    body.Click()
    // 3) Add a CSS class to the button:
    btn.Get("classList").Call("add", "active")
    // Inspect the class attribute
    cls := btn.Get("class").String()
    wantCls := "active"
    if cls != wantCls {
        t.Errorf("unexpected class attribute: got %s, want %s", cls, wantCls)
    }

    // Inspect the rendered HTML
    got := body.InnerHTML()
    want := "<button>Clicked</button><span>1</span>"
    if got != want {
        t.Errorf("unexpected HTML: got %s, want %s", got, want)
    }

    // For input-driven components, dispatch an InputMsg directly:
    send(foo.InputMsg{Value: "Alice"})
    got = body.InnerHTML()
    want = "Hello Alice"
    if got != want {
        t.Errorf("after send, HTML = %s; want %s", got, want)
    }
}
```

Then run:

```bash
go test ./example/...
```
