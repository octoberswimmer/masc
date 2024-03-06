package components

import (
	"github.com/octoberswimmer/rumtew"
	"github.com/octoberswimmer/rumtew/elem"
	"github.com/octoberswimmer/rumtew/event"
	"github.com/octoberswimmer/rumtew/prop"
)

// FilterButton is a rumtew.Component which allows the user to select a filter
// state.
type FilterButton struct {
	rumtew.Core

	Label        string      `vecty:"prop"`
	Filter       FilterState `vecty:"prop"`
	ActiveFilter bool
}

func (b *FilterButton) onClick(send func(rumtew.Msg)) func(*rumtew.Event) {
	return func(event *rumtew.Event) {
		send(SetFilter{Filter: b.Filter})
	}
}

// Render implements the vecty.Component interface.
func (b *FilterButton) Render(send func(rumtew.Msg)) rumtew.ComponentOrHTML {
	return elem.ListItem(
		elem.Anchor(
			rumtew.Markup(
				rumtew.MarkupIf(b.ActiveFilter, rumtew.Class("selected")),
				prop.Href("#"),
				event.Click(b.onClick(send)).PreventDefault(),
			),

			rumtew.Text(b.Label),
		),
	)
}

func (b *FilterButton) Copy() rumtew.Component {
	cpy := *b
	return &cpy
}
