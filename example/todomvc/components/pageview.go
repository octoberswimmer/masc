package components

import (
	"encoding/json"
	"fmt"
	"strconv"
	"syscall/js"

	"github.com/octoberswimmer/rumtew"
	"github.com/octoberswimmer/rumtew/elem"
	"github.com/octoberswimmer/rumtew/event"
	"github.com/octoberswimmer/rumtew/prop"
	"github.com/octoberswimmer/rumtew/style"
)

// PageView is a rumtew.Component which represents the entire page.
type PageView struct {
	rumtew.Core

	Items        []*Item `vecty:"prop"`
	newItemTitle string

	// Filter represents the active viewing filter for items.
	Filter FilterState
}

// Item represents a single Todo item in the store.
type Item struct {
	Title     string
	Completed bool
	editing   bool
	editTitle string
}

type ItemList []*Item

// FilterState represents a viewing filter for Todo items in the store.
type FilterState int

const (
	// All is a FilterState which shows all items.
	All FilterState = iota

	// Active is a FilterState which shows only non-completed items.
	Active

	// Completed is a FilterState which shows only completed items.
	Completed
)

func (m *PageView) Init() rumtew.Cmd {
	fmt.Println("Initializing PageView")
	return attachLocalStorage
}

func (m *PageView) Update(msg rumtew.Msg) (rumtew.Model, rumtew.Cmd) {
	var (
		cmds []rumtew.Cmd
	)
	switch msg := msg.(type) {
	case ItemList:
		m.Items = msg
		return m, nil
	case AddItemMsg:
		m.Items = append(m.Items, &Item{Title: m.newItemTitle})
		m.newItemTitle = ""
		cmds = append(cmds, m.updateLocalStorage)
	case NewItemTitleMsg:
		m.newItemTitle = msg.Title
	case ClearCompleted:
		var activeItems []*Item
		for _, item := range m.Items {
			if !item.Completed {
				activeItems = append(activeItems, item)
			}
		}
		m.Items = activeItems
		cmds = append(cmds, m.updateLocalStorage)
	case SetAllCompleted:
		for _, item := range m.Items {
			item.Completed = msg.Completed
		}
		cmds = append(cmds, m.updateLocalStorage)
	case SetFilter:
		m.Filter = msg.Filter
	case StartEdit:
		m.Items[msg.Index].editing = true
		m.Items[msg.Index].editTitle = m.Items[msg.Index].Title
	case EditInput:
		m.Items[msg.Index].editTitle = msg.Title
	case StopEdit:
		m.Items[msg.Index].editing = false
		m.Items[msg.Index].Title = m.Items[msg.Index].editTitle
		m.Items[msg.Index].editTitle = ""
		cmds = append(cmds, m.updateLocalStorage)
	case UpdateCompleted:
		m.Items[msg.Index].Completed = msg.Completed
		cmds = append(cmds, m.updateLocalStorage)
	case Destroy:
		m.Items = append(m.Items[:msg.Index], m.Items[msg.Index+1:]...)
		cmds = append(cmds, m.updateLocalStorage)
	}
	return m, rumtew.Batch(cmds...)
}

func (m *PageView) updateLocalStorage() rumtew.Msg {
	fmt.Printf("Marshalling %+v", m.Items)
	data, err := json.Marshal(m.Items)
	if err != nil {
		fmt.Println("failed to store items: " + err.Error())
	}
	js.Global().Get("localStorage").Set("items", string(data))
	fmt.Println("Updated local storage", string(data))
	return nil
}

func attachLocalStorage() rumtew.Msg {
	if data := js.Global().Get("localStorage").Get("items"); !data.IsUndefined() {
		fmt.Println("Got items from localStorage", data.String())
		var items ItemList
		if err := json.Unmarshal([]byte(data.String()), &items); err != nil {
			println("failed to load items: " + err.Error())
		}
		return items
	}
	return nil
}

func (p *PageView) onNewItemTitleInput(send func(rumtew.Msg)) func(*rumtew.Event) {
	return func(event *rumtew.Event) {
		// p.newItemTitle = event.Target.Get("value").String()
		send(NewItemTitleMsg{Title: event.Target.Get("value").String()})
	}
}

func (p *PageView) onAdd(send func(rumtew.Msg)) func(*rumtew.Event) {
	return func(event *rumtew.Event) {
		send(AddItemMsg{})
	}
}

func (p *PageView) onClearCompleted(send func(rumtew.Msg)) func(*rumtew.Event) {
	return func(event *rumtew.Event) {
		send(ClearCompleted{})
	}
}

func (p *PageView) onToggleAllCompleted(send func(rumtew.Msg)) func(*rumtew.Event) {
	return func(event *rumtew.Event) {
		send(SetAllCompleted{
			Completed: event.Target.Get("checked").Bool(),
		})
	}
}

// Render implements the rumtew.Component interface.
func (p *PageView) Render(send func(rumtew.Msg)) rumtew.ComponentOrHTML {
	return elem.Div(
		elem.Section(
			rumtew.Markup(
				rumtew.Class("todoapp"),
			),

			p.renderHeader(send),
			rumtew.If(len(p.Items) > 0,
				p.renderItemList(send),
				p.renderFooter(send),
			),
		),

		p.renderInfo(),
	)
}

func (p *PageView) renderHeader(send func(rumtew.Msg)) *rumtew.HTML {
	return elem.Header(
		rumtew.Markup(
			rumtew.Class("header"),
		),

		elem.Heading1(
			rumtew.Text("todos"),
		),
		elem.Form(
			rumtew.Markup(
				style.Margin(style.Px(0)),
				event.Submit(p.onAdd(send)).PreventDefault(),
			),

			elem.Input(
				rumtew.Markup(
					rumtew.Class("new-todo"),
					prop.Placeholder("What needs to be done?"),
					prop.Autofocus(true),
					prop.Value(p.newItemTitle),
					event.Input(p.onNewItemTitleInput(send)),
				),
			),
		),
	)
}

// ActiveItemCount returns the current number of items that are not completed.
func (p *PageView) ActiveItemCount() int {
	return p.count(false)
}

// CompletedItemCount returns the current number of items that are completed.
func (p *PageView) CompletedItemCount() int {
	return p.count(true)
}

func (p *PageView) count(completed bool) int {
	count := 0
	for _, item := range p.Items {
		if item.Completed == completed {
			count++
		}
	}
	return count
}

func (p *PageView) renderFooter(send func(rumtew.Msg)) *rumtew.HTML {
	count := p.ActiveItemCount()
	itemsLeftText := " items left"
	if count == 1 {
		itemsLeftText = " item left"
	}

	return elem.Footer(
		rumtew.Markup(
			rumtew.Class("footer"),
		),

		elem.Span(
			rumtew.Markup(
				rumtew.Class("todo-count"),
			),

			elem.Strong(
				rumtew.Text(strconv.Itoa(count)),
			),
			rumtew.Text(itemsLeftText),
		),

		elem.UnorderedList(
			rumtew.Markup(
				rumtew.Class("filters"),
			),
			&FilterButton{Label: "All", Filter: All, ActiveFilter: p.Filter == All},
			rumtew.Text(" "),
			&FilterButton{Label: "Active", Filter: Active, ActiveFilter: p.Filter == Active},
			rumtew.Text(" "),
			&FilterButton{Label: "Completed", Filter: Completed, ActiveFilter: p.Filter == Completed},
		),

		rumtew.If(p.CompletedItemCount() > 0,
			elem.Button(
				rumtew.Markup(
					rumtew.Class("clear-completed"),
					event.Click(p.onClearCompleted(send)),
				),
				rumtew.Text("Clear completed ("+strconv.Itoa(p.CompletedItemCount())+")"),
			),
		),
	)
}

func (p *PageView) renderInfo() *rumtew.HTML {
	return elem.Footer(
		rumtew.Markup(
			rumtew.Class("info"),
		),

		elem.Paragraph(
			rumtew.Text("Double-click to edit a todo"),
		),
		elem.Paragraph(
			rumtew.Text("Created by "),
			elem.Anchor(
				rumtew.Markup(
					prop.Href("http://github.com/neelance"),
				),
				rumtew.Text("Richard Musiol"),
			),
		),
		elem.Paragraph(
			rumtew.Text("Part of "),
			elem.Anchor(
				rumtew.Markup(
					prop.Href("http://todomvc.com"),
				),
				rumtew.Text("TodoMVC"),
			),
		),
	)
}

func (p *PageView) renderItemList(send func(rumtew.Msg)) *rumtew.HTML {
	var items rumtew.List
	for i, item := range p.Items {
		if (p.Filter == Active && item.Completed) || (p.Filter == Completed && !item.Completed) {
			continue
		}
		iv := &ItemView{Index: i, Title: item.Title, Completed: item.Completed, Editing: item.editing}
		if iv.Editing {
			iv.EditTitle = item.editTitle
		}

		items = append(items, iv)
	}

	return elem.Section(
		rumtew.Markup(
			rumtew.Class("main"),
		),

		elem.Input(
			rumtew.Markup(
				rumtew.Class("toggle-all"),
				prop.ID("toggle-all"),
				prop.Type(prop.TypeCheckbox),
				prop.Checked(p.CompletedItemCount() == len(p.Items)),
				event.Change(p.onToggleAllCompleted(send)),
			),
		),
		elem.Label(
			rumtew.Markup(
				prop.For("toggle-all"),
			),
			rumtew.Text("Mark all as complete"),
		),

		elem.UnorderedList(
			rumtew.Markup(
				rumtew.Class("todo-list"),
			),
			items,
		),
	)
}

func (p *PageView) Copy() rumtew.Component {
	cpy := *p
	return &cpy
}
