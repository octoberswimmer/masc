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

[embedmd]:# (example/hellovecty/hellovecty.go)
```go
package main

import (
	"github.com/octoberswimmer/masc"
	"github.com/octoberswimmer/masc/elem"
)

func main() {
	masc.SetTitle("Hello Vecty!")
	m := &PageView{}
	pgm := masc.NewProgram(m)
	_, err := pgm.Run()
	if err != nil {
		panic(err)
	}
}

// PageView is our main page component.
type PageView struct {
	masc.Core
}

func (p *PageView) Init() masc.Cmd {
	return nil
}

func (p *PageView) Update(msg masc.Msg) (masc.Model, masc.Cmd) {
	return p, nil
}

// Render implements the masc.Component interface.
func (p *PageView) Render(send func(masc.Msg)) masc.ComponentOrHTML {
	return elem.Body(
		masc.Text("Hello Vecty!"),
	)
}
```

Additional examples, including a todo app,
are in the [example](example/) directory.  These can be run using
[wasmserve](https://github.com/hajimehoshi/wasmserve).
