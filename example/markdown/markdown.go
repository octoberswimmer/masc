package main

import (
	"bytes"

	"github.com/octoberswimmer/masc"
	"github.com/octoberswimmer/masc/elem"
	"github.com/octoberswimmer/masc/event"
	"github.com/yuin/goldmark"
)

func main() {
	masc.SetTitle("Markdown Demo")
	m := &PageView{
		Input: `# Markdown Example

This is a live editor, try editing the Markdown on the right of the page.
`,
	}
	pgm := masc.NewProgram(m)
	_, err := pgm.Run()
	if err != nil {
		panic(err)
	}
}

type UpdateInput struct {
	Input string
}

// PageView is our main page component.
type PageView struct {
	masc.Core
	Input string
}

func (p *PageView) Init() masc.Cmd {
	return nil
}

func (p *PageView) Update(msg masc.Msg) (masc.Model, masc.Cmd) {
	switch msg := msg.(type) {
	case UpdateInput:
		p.Input = msg.Input
	}
	return p, nil
}

// Render implements the masc.Component interface.
func (p *PageView) Render(send func(masc.Msg)) masc.ComponentOrHTML {
	return elem.Body(
		// Display a textarea on the right-hand side of the page.
		elem.Div(
			masc.Markup(
				masc.Style("float", "right"),
			),
			elem.TextArea(
				masc.Markup(
					masc.Style("font-family", "monospace"),
					masc.Property("rows", 14),
					masc.Property("cols", 70),

					// When input is typed into the textarea, update the local
					// component state and rerender.
					event.Input(func(e *masc.Event) {
						v := e.Target.Get("value").String()
						send(UpdateInput{v})
					}),
				),
				masc.Text(p.Input), // initial textarea text.
			),
		),

		// Render the markdown.
		&Markdown{Input: p.Input},
	)
}

// Markdown is a simple component which renders the Input markdown as sanitized
// HTML into a div.
type Markdown struct {
	masc.Core
	Input string `masc:"prop"`
}

// Render implements the masc.Component interface.
func (m *Markdown) Render(send func(masc.Msg)) masc.ComponentOrHTML {
	// Render the markdown input into HTML using Goldmark.
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(m.Input), &buf); err != nil {
		panic(err)
	}
	// The goldmark README says:
	// "By default, goldmark does not render raw HTML or potentially dangerous links. "
	// So, it should be ok without sanitizing.

	// Return the HTML.
	return elem.Div(
		masc.Markup(
			masc.UnsafeHTML(buf.String()),
		),
	)
}
