package components

import (
	"github.com/octoberswimmer/masc"
	"github.com/octoberswimmer/masc/elem"
	"github.com/octoberswimmer/masc/event"
	"github.com/octoberswimmer/masc/prop"
)

// FilterButton is a masc.Component which allows the user to select a filter
// state.
type FilterButton struct {
	masc.Core

	Label        string      `masc:"prop"`
	Filter       FilterState `masc:"prop"`
	ActiveFilter bool
}

func (b *FilterButton) onClick(send func(masc.Msg)) func(*masc.Event) {
	return func(event *masc.Event) {
		send(SetFilter{Filter: b.Filter})
	}
}

// Render implements the masc.Component interface.
func (b *FilterButton) Render(send func(masc.Msg)) masc.ComponentOrHTML {
	return elem.ListItem(
		elem.Anchor(
			masc.Markup(
				masc.MarkupIf(b.ActiveFilter, masc.Class("selected")),
				prop.Href("#"),
				event.Click(b.onClick(send)).PreventDefault(),
			),

			masc.Text(b.Label),
		),
	)
}

func (b *FilterButton) Copy() masc.Component {
	cpy := *b
	return &cpy
}
