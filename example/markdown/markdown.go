package main

import (
	"bytes"

	"github.com/octoberswimmer/rumtew"
	"github.com/octoberswimmer/rumtew/elem"
	"github.com/octoberswimmer/rumtew/event"
	"github.com/yuin/goldmark"
)

func main() {
	rumtew.SetTitle("Markdown Demo")
	m := &PageView{
		Input: `# Markdown Example

This is a live editor, try editing the Markdown on the right of the page.
`,
	}
	pgm := rumtew.NewProgram(m)
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
	rumtew.Core
	Input string
}

func (p *PageView) Init() rumtew.Cmd {
	return nil
}

func (p *PageView) Update(msg rumtew.Msg) (rumtew.Model, rumtew.Cmd) {
	switch msg := msg.(type) {
	case UpdateInput:
		p.Input = msg.Input
	}
	return p, nil
}

// Render implements the rumtew.Component interface.
func (p *PageView) Render(send func(rumtew.Msg)) rumtew.ComponentOrHTML {
	return elem.Body(
		// Display a textarea on the right-hand side of the page.
		elem.Div(
			rumtew.Markup(
				rumtew.Style("float", "right"),
			),
			elem.TextArea(
				rumtew.Markup(
					rumtew.Style("font-family", "monospace"),
					rumtew.Property("rows", 14),
					rumtew.Property("cols", 70),

					// When input is typed into the textarea, update the local
					// component state and rerender.
					event.Input(func(e *rumtew.Event) {
						v := e.Target.Get("value").String()
						send(UpdateInput{v})
					}),
				),
				rumtew.Text(p.Input), // initial textarea text.
			),
		),

		// Render the markdown.
		&Markdown{Input: p.Input},
	)
}

// Markdown is a simple component which renders the Input markdown as sanitized
// HTML into a div.
type Markdown struct {
	rumtew.Core
	Input string `vecty:"prop"`
}

// Render implements the rumtew.Component interface.
func (m *Markdown) Render(send func(rumtew.Msg)) rumtew.ComponentOrHTML {
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
		rumtew.Markup(
			rumtew.UnsafeHTML(buf.String()),
		),
	)
}
