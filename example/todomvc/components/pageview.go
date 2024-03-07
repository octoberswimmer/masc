package components

import (
	"encoding/json"
	"fmt"
	"strconv"
	"syscall/js"

	"github.com/octoberswimmer/masc"
	"github.com/octoberswimmer/masc/elem"
	"github.com/octoberswimmer/masc/event"
	"github.com/octoberswimmer/masc/prop"
	"github.com/octoberswimmer/masc/style"
)

// PageView is a masc.Component which represents the entire page.
type PageView struct {
	masc.Core

	Items        []*Item `masc:"prop"`
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

func (m *PageView) Init() masc.Cmd {
	fmt.Println("Initializing PageView")
	return attachLocalStorage
}

func (m *PageView) Update(msg masc.Msg) (masc.Model, masc.Cmd) {
	var (
		cmds []masc.Cmd
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
	return m, masc.Batch(cmds...)
}

func (m *PageView) updateLocalStorage() masc.Msg {
	fmt.Printf("Marshalling %+v", m.Items)
	data, err := json.Marshal(m.Items)
	if err != nil {
		fmt.Println("failed to store items: " + err.Error())
	}
	js.Global().Get("localStorage").Set("items", string(data))
	fmt.Println("Updated local storage", string(data))
	return nil
}

func attachLocalStorage() masc.Msg {
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

func (p *PageView) onNewItemTitleInput(send func(masc.Msg)) func(*masc.Event) {
	return func(event *masc.Event) {
		// p.newItemTitle = event.Target.Get("value").String()
		send(NewItemTitleMsg{Title: event.Target.Get("value").String()})
	}
}

func (p *PageView) onAdd(send func(masc.Msg)) func(*masc.Event) {
	return func(event *masc.Event) {
		send(AddItemMsg{})
	}
}

func (p *PageView) onClearCompleted(send func(masc.Msg)) func(*masc.Event) {
	return func(event *masc.Event) {
		send(ClearCompleted{})
	}
}

func (p *PageView) onToggleAllCompleted(send func(masc.Msg)) func(*masc.Event) {
	return func(event *masc.Event) {
		send(SetAllCompleted{
			Completed: event.Target.Get("checked").Bool(),
		})
	}
}

// Render implements the masc.Component interface.
func (p *PageView) Render(send func(masc.Msg)) masc.ComponentOrHTML {
	return elem.Div(
		elem.Section(
			masc.Markup(
				masc.Class("todoapp"),
			),

			p.renderHeader(send),
			masc.If(len(p.Items) > 0,
				p.renderItemList(send),
				p.renderFooter(send),
			),
		),

		p.renderInfo(),
	)
}

func (p *PageView) renderHeader(send func(masc.Msg)) *masc.HTML {
	return elem.Header(
		masc.Markup(
			masc.Class("header"),
		),

		elem.Heading1(
			masc.Text("todos"),
		),
		elem.Form(
			masc.Markup(
				style.Margin(style.Px(0)),
				event.Submit(p.onAdd(send)).PreventDefault(),
			),

			elem.Input(
				masc.Markup(
					masc.Class("new-todo"),
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

func (p *PageView) renderFooter(send func(masc.Msg)) *masc.HTML {
	count := p.ActiveItemCount()
	itemsLeftText := " items left"
	if count == 1 {
		itemsLeftText = " item left"
	}

	return elem.Footer(
		masc.Markup(
			masc.Class("footer"),
		),

		elem.Span(
			masc.Markup(
				masc.Class("todo-count"),
			),

			elem.Strong(
				masc.Text(strconv.Itoa(count)),
			),
			masc.Text(itemsLeftText),
		),

		elem.UnorderedList(
			masc.Markup(
				masc.Class("filters"),
			),
			&FilterButton{Label: "All", Filter: All, ActiveFilter: p.Filter == All},
			masc.Text(" "),
			&FilterButton{Label: "Active", Filter: Active, ActiveFilter: p.Filter == Active},
			masc.Text(" "),
			&FilterButton{Label: "Completed", Filter: Completed, ActiveFilter: p.Filter == Completed},
		),

		masc.If(p.CompletedItemCount() > 0,
			elem.Button(
				masc.Markup(
					masc.Class("clear-completed"),
					event.Click(p.onClearCompleted(send)),
				),
				masc.Text("Clear completed ("+strconv.Itoa(p.CompletedItemCount())+")"),
			),
		),
	)
}

func (p *PageView) renderInfo() *masc.HTML {
	return elem.Footer(
		masc.Markup(
			masc.Class("info"),
		),

		elem.Paragraph(
			masc.Text("Double-click to edit a todo"),
		),
		elem.Paragraph(
			masc.Text("Created by "),
			elem.Anchor(
				masc.Markup(
					prop.Href("http://github.com/neelance"),
				),
				masc.Text("Richard Musiol"),
			),
		),
		elem.Paragraph(
			masc.Text("Part of "),
			elem.Anchor(
				masc.Markup(
					prop.Href("http://todomvc.com"),
				),
				masc.Text("TodoMVC"),
			),
		),
	)
}

func (p *PageView) renderItemList(send func(masc.Msg)) *masc.HTML {
	var items masc.List
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
		masc.Markup(
			masc.Class("main"),
		),

		elem.Input(
			masc.Markup(
				masc.Class("toggle-all"),
				prop.ID("toggle-all"),
				prop.Type(prop.TypeCheckbox),
				prop.Checked(p.CompletedItemCount() == len(p.Items)),
				event.Change(p.onToggleAllCompleted(send)),
			),
		),
		elem.Label(
			masc.Markup(
				prop.For("toggle-all"),
			),
			masc.Text("Mark all as complete"),
		),

		elem.UnorderedList(
			masc.Markup(
				masc.Class("todo-list"),
			),
			items,
		),
	)
}

func (p *PageView) Copy() masc.Component {
	cpy := *p
	return &cpy
}
