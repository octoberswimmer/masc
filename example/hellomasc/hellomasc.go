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
					// Grab the new value from the event target element
					v := e.Target.Get("value").String()
					send(InputMsg{Value: v})
				}),
			),
		),
		elem.Break(),
		// Greeting text includes both input and clicks
		masc.Text("Hello "+p.input+""+strings.Repeat("!", p.clicks)),
	)
}
