package components

import (
	"fmt"

	"github.com/octoberswimmer/masc"
	"github.com/octoberswimmer/masc/elem"
	"github.com/octoberswimmer/masc/event"
	"github.com/octoberswimmer/masc/prop"
	"github.com/octoberswimmer/masc/style"
)

// ItemView is a masc.Component which represents a single item in the TODO
// list.
type ItemView struct {
	masc.Core

	Index     int    `masc:"prop"`
	Title     string `masc:"prop"`
	Completed bool   `masc:"prop"`

	Editing   bool   `masc:"prop"`
	EditTitle string `masc:"prop"`
	input     *masc.HTML
}

// Key implements the masc.Keyer interface.
func (p *ItemView) Key() interface{} {
	return p.Index
}

type StartEdit struct {
	Index int
}

type Destroy struct {
	Index int
}

type UpdateCompleted struct {
	Index     int
	Completed bool
}

type StopEdit struct {
	Index int
}

type EditInput struct {
	Index int
	Title string
}

func (p *ItemView) onDestroy(send func(masc.Msg)) func(*masc.Event) {
	return func(event *masc.Event) {
		send(Destroy{p.Index})
	}
}

func (p *ItemView) onToggleCompleted(send func(masc.Msg)) func(*masc.Event) {
	return func(event *masc.Event) {
		send(UpdateCompleted{
			Index:     p.Index,
			Completed: event.Target.Get("checked").Bool(),
		})
	}
}

func (p *ItemView) onStartEdit(send func(masc.Msg)) func(*masc.Event) {
	return func(event *masc.Event) {
		send(StartEdit{p.Index})
	}
}

func (p *ItemView) onEditInput(send func(masc.Msg)) func(*masc.Event) {
	return func(event *masc.Event) {
		send(EditInput{p.Index, event.Target.Get("value").String()})
	}
}

func (p *ItemView) onStopEdit(send func(masc.Msg)) func(*masc.Event) {
	return func(event *masc.Event) {
		send(StopEdit{p.Index})
	}
}

// Render implements the masc.Component interface.
func (p *ItemView) Render(send func(masc.Msg)) masc.ComponentOrHTML {
	p.input = elem.Input(
		masc.Markup(
			masc.Class("edit"),
			masc.Attribute("id", fmt.Sprintf("input-%d", p.Index)),
			prop.Value(p.EditTitle),
			event.Input(p.onEditInput(send)),
		),
	)

	return elem.ListItem(
		masc.Markup(
			masc.ClassMap{
				"completed": p.Completed,
				"editing":   p.Editing,
			},
		),

		elem.Div(
			masc.Markup(
				masc.Class("view"),
			),

			elem.Input(
				masc.Markup(
					masc.Class("toggle"),
					prop.Type(prop.TypeCheckbox),
					prop.Checked(p.Completed),
					event.Change(p.onToggleCompleted(send)),
				),
			),
			elem.Label(
				masc.Markup(
					event.DoubleClick(p.onStartEdit(send)),
				),
				masc.Text(p.Title),
			),
			elem.Button(
				masc.Markup(
					masc.Class("destroy"),
					event.Click(p.onDestroy(send)),
				),
			),
		),
		elem.Form(
			masc.Markup(
				style.Margin(style.Px(0)),
				event.Submit(p.onStopEdit(send)).PreventDefault(),
			),
			p.input,
		),
	)
}

func (p *ItemView) Copy() masc.Component {
	cpy := *p
	return &cpy
}
