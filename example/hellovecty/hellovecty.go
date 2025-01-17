package main

import (
	"strings"

	"github.com/octoberswimmer/masc"
	"github.com/octoberswimmer/masc/elem"
	"github.com/octoberswimmer/masc/event"
)

func main() {
	masc.SetTitle("Hello masc User!")
	m := &PageView{}
	pgm := masc.NewProgram(m)
	_, err := pgm.Run()
	if err != nil {
		panic(err)
	}
}

type ClickMsg struct{}

// PageView is our main page component.
type PageView struct {
	masc.Core

	// The model state is the number of clicks
	clicks int
}

func (p *PageView) Init() masc.Cmd {
	return nil
}

// PageView is a masc.Model; it has Init and Update functions
func (p *PageView) Update(msg masc.Msg) (masc.Model, masc.Cmd) {
	switch msg.(type) {
	case ClickMsg:
		// Update the model state when we get a click message
		p.clicks++
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
		masc.Text("Hello masc User"+strings.Repeat("!", p.clicks)),
	)
}
