package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/hexops/vecty"
	"github.com/hexops/vecty/example/todomvc/actions"
	"github.com/hexops/vecty/example/todomvc/components"
	"github.com/hexops/vecty/example/todomvc/dispatcher"
	"github.com/hexops/vecty/example/todomvc/store"
	"github.com/hexops/vecty/example/todomvc/store/model"
)

func main() {
	attachLocalStorage()

	vecty.SetTitle("GopherJS • TodoMVC")
	vecty.AddStylesheet("https://rawgit.com/tastejs/todomvc-common/master/base.css")
	vecty.AddStylesheet("https://rawgit.com/tastejs/todomvc-app-css/master/index.css")
	p := &components.PageView{}
	store.Listeners.Add(p, func() {
		p.Items = store.Items
		vecty.Rerender(p)
	})
	vecty.RenderBody(p)
}

func attachLocalStorage() {
	store.Listeners.Add(nil, func() {
		data, err := json.Marshal(store.Items)
		if err != nil {
			println("failed to store items: " + err.Error())
		}
		js.Global().Get("localStorage").Set("items", string(data))
	})

	if data := js.Global().Get("localStorage").Get("items"); !data.IsUndefined() {
		var items []*model.Item
		if err := json.Unmarshal([]byte(data.String()), &items); err != nil {
			println("failed to load items: " + err.Error())
		}
		dispatcher.Dispatch(&actions.ReplaceItems{
			Items: items,
		})
	}
}
