package components

import (
	"fmt"

	"github.com/octoberswimmer/rumtew"
	"github.com/octoberswimmer/rumtew/elem"
	"github.com/octoberswimmer/rumtew/event"
	"github.com/octoberswimmer/rumtew/prop"
	"github.com/octoberswimmer/rumtew/style"
)

// ItemView is a rumtew.Component which represents a single item in the TODO
// list.
type ItemView struct {
	rumtew.Core

	Index     int    `vecty:"prop"`
	Title     string `vecty:"prop"`
	Completed bool   `vecty:"prop"`

	Editing   bool   `vecty:"prop"`
	EditTitle string `vecty:"prop"`
	input     *rumtew.HTML
}

// Key implements the rumtew.Keyer interface.
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

func (p *ItemView) onDestroy(send func(rumtew.Msg)) func(*rumtew.Event) {
	return func(event *rumtew.Event) {
		send(Destroy{p.Index})
	}
}

func (p *ItemView) onToggleCompleted(send func(rumtew.Msg)) func(*rumtew.Event) {
	return func(event *rumtew.Event) {
		send(UpdateCompleted{
			Index:     p.Index,
			Completed: event.Target.Get("checked").Bool(),
		})
	}
}

func (p *ItemView) onStartEdit(send func(rumtew.Msg)) func(*rumtew.Event) {
	return func(event *rumtew.Event) {
		send(StartEdit{p.Index})
	}
}

func (p *ItemView) onEditInput(send func(rumtew.Msg)) func(*rumtew.Event) {
	return func(event *rumtew.Event) {
		send(EditInput{p.Index, event.Target.Get("value").String()})
	}
}

func (p *ItemView) onStopEdit(send func(rumtew.Msg)) func(*rumtew.Event) {
	return func(event *rumtew.Event) {
		send(StopEdit{p.Index})
	}
}

// Render implements the rumtew.Component interface.
func (p *ItemView) Render(send func(rumtew.Msg)) rumtew.ComponentOrHTML {
	p.input = elem.Input(
		rumtew.Markup(
			rumtew.Class("edit"),
			rumtew.Attribute("id", fmt.Sprintf("input-%d", p.Index)),
			prop.Value(p.EditTitle),
			event.Input(p.onEditInput(send)),
		),
	)

	return elem.ListItem(
		rumtew.Markup(
			rumtew.ClassMap{
				"completed": p.Completed,
				"editing":   p.Editing,
			},
		),

		elem.Div(
			rumtew.Markup(
				rumtew.Class("view"),
			),

			elem.Input(
				rumtew.Markup(
					rumtew.Class("toggle"),
					prop.Type(prop.TypeCheckbox),
					prop.Checked(p.Completed),
					event.Change(p.onToggleCompleted(send)),
				),
			),
			elem.Label(
				rumtew.Markup(
					event.DoubleClick(p.onStartEdit(send)),
				),
				rumtew.Text(p.Title),
			),
			elem.Button(
				rumtew.Markup(
					rumtew.Class("destroy"),
					event.Click(p.onDestroy(send)),
				),
			),
		),
		elem.Form(
			rumtew.Markup(
				style.Margin(style.Px(0)),
				event.Submit(p.onStopEdit(send)).PreventDefault(),
			),
			p.input,
		),
	)
}

func (p *ItemView) Copy() rumtew.Component {
	cpy := *p
	return &cpy
}
